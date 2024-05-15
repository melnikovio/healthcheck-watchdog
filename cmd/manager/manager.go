package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/healthcheck-watchdog/cmd/authentication"
	"github.com/healthcheck-watchdog/cmd/clients"
	httpclient "github.com/healthcheck-watchdog/cmd/clients/http"
	k8sclient "github.com/healthcheck-watchdog/cmd/clients/kubernetes"
	wsclient "github.com/healthcheck-watchdog/cmd/clients/websocket"
	"github.com/healthcheck-watchdog/cmd/exporter"
	"github.com/healthcheck-watchdog/cmd/model"
	"github.com/healthcheck-watchdog/cmd/watchdog"

	log "github.com/sirupsen/logrus"
)

// Task executor module
type Manager struct {
	exporter  *exporter.Exporter
	watchdog  *watchdog.WatchDog
	config    *model.Config
	executors map[string]clients.Executor
	Jobs      map[string]*model.TaskStatus
	mutex     sync.Mutex
}

// Launch task executor
func Start(exporter *exporter.Exporter, config *model.Config) {
	NewManager(exporter, config)
}

func NewManager(exporter *exporter.Exporter, config *model.Config) *Manager {
	executor := &Manager{
		config:    config,
		executors: make(map[string]clients.Executor),
		Jobs:      make(map[string]*model.TaskStatus),
		exporter:  exporter,
	}

	authClient := authentication.NewAuthClient(config)

	httpClient :=
		httpclient.NewHttpClient(authClient, config)
	websocketClient :=
		wsclient.NewWsClient(authClient, config)
	kubernetesClient, err :=
		k8sclient.NewKubernetesClient(config)
	if err != nil {
		executor.executors["memory"] = kubernetesClient
		executor.watchdog = watchdog.NewWatchDog(kubernetesClient, config)
	}

	executor.executors["http_get"] = httpClient
	executor.executors["http_post"] = httpClient
	executor.executors["websocket"] = websocketClient

	go executor.run()

	return executor
}

// Get task
func (m *Manager) GetTask(key string) *model.TaskStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Jobs[key]
}

// Update task
func (m *Manager) CreateOrUpdateTask(id string, job *model.Job, updater func(*model.TaskStatus)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, ok := m.Jobs[id]
	if !ok {
		task = &model.TaskStatus{
			Id: id,
		}
		if job != nil {
			task.Job = job
		}

		m.Jobs[id] = task
	}

	updater(task)
}

// Runner
func (m *Manager) run() {
	// Channel for task results
	resultChan := make(chan *model.TaskResult)
	go m.resultProcessor(resultChan)

	// Infinite runner
	for {
		for i := range m.config.Jobs {
			if m.isTaskShoudRun(&m.config.Jobs[i]) {
				go m.processTask(&m.config.Jobs[i], resultChan)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// Check if task should run
func (m *Manager) isTaskShoudRun(job *model.Job) bool {
	task := m.GetTask(job.Id)

	// check if currently running
	if task != nil && task.Running {
		return false
	}

	// check dependent job and return false if not running
	if job.DependentJob != "" {
		dependentJob := m.GetTask(job.DependentJob)
		if dependentJob == nil || !dependentJob.Status {
			log.Info(fmt.Sprintf("%s Dependent job %s not ready", job.Id, job.DependentJob))
			return false
		}
	}

	// if task never run - allow to run
	if task == nil {
		return true
	}

	// if task alreay run, then check timeout
	diff := time.Now().Unix() - task.LastCall
	return diff >= int64(job.Timeout)
}

// Process task to executors
func (m *Manager) processTask(job *model.Job, resultChan chan *model.TaskResult) {
	// Searching for clients
	client, ok := m.executors[job.Type]
	if !ok {
		log.Error(fmt.Sprintf("Client for job type %s not found", job.Type))
		return
	}

	// Update task
	m.CreateOrUpdateTask(job.Id, job, func(task *model.TaskStatus) {
		task.Running = true
	})

	// Executing task with repeated delay
	client.Execute(job, resultChan)
}

// Process results from tasks
func (m *Manager) resultProcessor(resultChan <-chan *model.TaskResult) {
	for result := range resultChan {
		log.Info(fmt.Sprintf("Manager: Processed result %v", result))

		m.CreateOrUpdateTask(result.Id, nil, func(task *model.TaskStatus) {
			task.Running = result.Running
			task.LastCall = time.Now().Unix()
			task.Status = result.Result
			task.LastResult = result

		})

		if m.exporter != nil {
			m.exporter.Channel <- *m.GetTask(result.Id)
		}

		if m.watchdog != nil {
			m.watchdog.Channel <- *m.GetTask(result.Id)
		}
	}
}

// Get readiness for healthcheck
func (m *Manager) Ready() (bool, error) {
	return m.config != nil &&
		len(m.executors) > 0 &&
		m.exporter == nil, nil
}

// Get liveness for healthcheck
func (m *Manager) Live() (bool, error) {
	return len(m.Jobs) > 0, nil
}
