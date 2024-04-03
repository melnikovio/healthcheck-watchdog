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
	exporter *exporter.Exporter
	watchdog *watchdog.WatchDog
	config *model.Config
	executors map[string]clients.Executor
	jobs map[string]*model.SchedulerTask
	mutex sync.Mutex
}

// Launch task executor 
func Start(exporter *exporter.Exporter, 
	watchdog *watchdog.WatchDog, config *model.Config) {
		NewManager(exporter, watchdog, config)
}

func NewManager(exporter *exporter.Exporter, 
	watchdog *watchdog.WatchDog, config *model.Config) *Manager {
	executor := &Manager{
		config: config,
		executors: make(map[string]clients.Executor),
		jobs: make(map[string]*model.SchedulerTask),
		exporter: exporter,
		watchdog: watchdog,
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
	}

	executor.executors["http_get"] = httpClient
	executor.executors["http_post"] = httpClient
	executor.executors["websocket"] = websocketClient

	executor.run()

	return executor
}

// Get task
func (m *Manager) getTask(key string) *model.SchedulerTask {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.jobs[key]
}

// Update task
func (m *Manager) createOrUpdateTask(key string, updater func(*model.SchedulerTask)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, ok := m.jobs[key]
	if !ok {
		task = &model.SchedulerTask{
			Id: key,
		}
		m.jobs[key] = task
	}	

	updater(task)
}

func (m *Manager) run() {
	// Channel for task results
	resultChan := make(chan *model.TaskResult)
	go m.resultProcessor(resultChan)

	for {
		for i := range m.config.Jobs {
			if m.isTaskShoudRun(&m.config.Jobs[i]) {
				go m.processTask(&m.config.Jobs[i], resultChan)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) isTaskShoudRun(job *model.Job) bool {
	task := m.getTask(job.Id)

	// check if currently running
	if task != nil && task.Running {
		return false
	}

	// check dependent job and return false if not running
	if job.DependentJob != "" {
		dependentJob := m.getTask(job.DependentJob)
		if dependentJob == nil || !dependentJob.LastStatus {
			log.Info("Dependent job not ready")
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

// Process task to clients
func (m *Manager) processTask(job *model.Job, resultChan chan *model.TaskResult) {
	// Searching for clients
	client, ok := m.executors[job.Type]
	if !ok {
		log.Error(fmt.Sprintf("Client for job type %s not found", job.Type))
		return
	}

	// Update task
	m.createOrUpdateTask(job.Id, func(task *model.SchedulerTask) {
		task.Running = true
	})

	// Executing task with repeated delay
	client.Execute(job, resultChan)
}

// Process results from tasks
func (m *Manager) resultProcessor(resultChan <-chan *model.TaskResult) {
	for result := range resultChan {
		log.Info(fmt.Sprintf("Manager: Processed result %v", result))

		m.createOrUpdateTask(result.Id, func(task *model.SchedulerTask) {
			task.Running = false
			task.LastCall = time.Now().Unix()
			task.LastStatus = result.Result
		})

		if m.exporter != nil {
			m.exporter.Channel<- result
		}
		if m.watchdog != nil {
			m.watchdog.Channel<- result
		}
	}
}

func (m *Manager) Status() (bool, error) {
	return true, nil
}