package watchdog

import (
	"fmt"
	"sync"
	"time"

	clients "github.com/healthcheck-watchdog/cmd/clients/kubernetes"
	redis "github.com/healthcheck-watchdog/cmd/clients/redis"
	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
)

type Fail struct {
	// required: true
	FailureChecks int `json:"failure_checks,omitempty"`
	// required: true
	RestartTime int64 `json:"restartTime,omitempty"`
}

type WatchDog struct {
	cluster *clients.KubernetesClient
	redis   *redis.Redis
	config  *model.Config
	mu      sync.RWMutex
	fails   map[string]Fail
	Channel chan model.TaskStatus
}

func NewWatchDog(k8sclient *clients.KubernetesClient, config *model.Config) *WatchDog {
	if config.WatchDog.Namespace == "" &&
		len(config.WatchDog.Actions) == 0 {
		log.Info("Missing watchdog configuration. Watchdog configuration ignored.")
		return nil
	}

	wd := WatchDog{
		cluster: k8sclient,
		redis:   redis.NewRedis(),
		Channel: make(chan model.TaskStatus, len(config.Jobs)),
		config:  config,
		fails:   make(map[string]Fail),
	}

	// Channel for task results
	go wd.resultProcessor(wd.Channel)

	return &wd
}

// Process results from tasks
func (ws *WatchDog) resultProcessor(resultChan <-chan model.TaskStatus) {
	for result := range resultChan {
		if result.Job == nil || !result.Job.WatchDogAction.Enabled {
			return
		}

		log.Trace(fmt.Sprintf("Watchdog: Processed result %v", result))

		ws.setStatus(&result)

		if ws.isWatchdogShoudRun(&result) {
			log.Info(fmt.Sprintf("Watchdog: Started watchdog %v", result))
			ws.Execute(result.Job)
		}
	}
}

// Set status
func (ws *WatchDog) setStatus(task *model.TaskStatus) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	status, ok := ws.fails[task.Job.Id]
	if !ok {
		status = Fail{}
	}

	if task.Status {
		status.FailureChecks = 0
	} else {
		status.FailureChecks++
	}

	ws.fails[task.Job.Id] = status
}

// Set status
func (ws *WatchDog) setTime(id string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	status, ok := ws.fails[id]
	if ok {
		status.RestartTime = time.Now().Unix()
	}
}

// Set status
func (ws *WatchDog) getStatus(id string) *Fail {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	status, ok := ws.fails[id]
	if !ok {
		status = Fail{}
	}

	return &status
}

// Check if watchdog should run
func (ws *WatchDog) isWatchdogShoudRun(task *model.TaskStatus) bool {
	if !task.Job.WatchDogAction.Enabled {
		return false
	}

	return ws.getStatus(task.Job.Id).FailureChecks >= task.Job.WatchDogAction.FailureThreshold &&
		(time.Now().Unix()-ws.getStatus(task.Job.Id).RestartTime) > task.Job.WatchDogAction.AwaitAfterRestart
}

func (ws *WatchDog) Execute(job *model.Job) {
	log.Info(fmt.Sprintf("Started watchdog actions: %v", job.Id))
	
	ws.setTime(job.Id)

	for _, jobAction := range job.WatchDogAction.Actions {
		for _, action := range ws.config.WatchDog.Actions {
			var err error
			if action.Id == jobAction {
				switch action.Type {
				case "redis":
					err = ws.redis.Execute(action.ConnectionString, action.Cmd)
				case "deployment_scale_down":
					err = ws.cluster.ScaleDown(action.Items, ws.config.WatchDog.Namespace)
				case "deployment_scale_up":
					err = ws.cluster.ScaleUp(action.Items, ws.config.WatchDog.Namespace)
				}
			}

			if err != nil {
				log.Error(fmt.Sprintf("Error in task %s: %s", action.Id, err.Error()))
			}
		}
	}
}
