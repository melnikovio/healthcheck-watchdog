package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/healthcheck-watchdog/cmd/authentication"
	"github.com/healthcheck-watchdog/cmd/cluster"
	"github.com/healthcheck-watchdog/cmd/exporter"
	"github.com/healthcheck-watchdog/cmd/model"
	"github.com/healthcheck-watchdog/cmd/watchdog"
	log "github.com/sirupsen/logrus"
)

type HealthCheck struct {
	config     *model.Config
	authClient *authentication.AuthClient
	status     *model.Status
	wsClient   *WsClient
	exporter   *exporter.Exporter
	watchDog   *watchdog.WatchDog
	httpClient *http.Client
	cluster    *cluster.Cluster
}

func NewHealthCheck(config *model.Config, authClient *authentication.AuthClient, ex *exporter.Exporter, wd *watchdog.WatchDog, cl *cluster.Cluster) *HealthCheck {
	hc := HealthCheck{
		config:     config,
		authClient: authClient,
		status: &model.Status{
			Tasks: make(map[string]*model.Task),
		},
		wsClient:   NewWsClient(ex),
		exporter:   ex,
		watchDog:   wd,
		httpClient: &http.Client{},
		cluster:    cl,
	}

	hc.InitTasks()

	return &hc
}

func (hc *HealthCheck) InitTasks() {
	for i := 0; i < len(hc.config.Jobs); i++ {
		hc.InitTask(&hc.config.Jobs[i])
	}

	for i := 0; i < len(hc.config.Jobs); i++ {
		go hc.StartTask(&hc.config.Jobs[i])
	}
}

func (hc *HealthCheck) getTask(taskId string) *model.Task {
	task := hc.getTaskFromMap(taskId)
	if task == nil {
		task = &model.Task{
			Id:            taskId,
			Online:        false,
			SuccessChecks: 0,
			FailureChecks: 0,
			RestartTime:   0,
		}
		hc.setTaskInMap(taskId, task)
	}

	return task
}

func (hc *HealthCheck) getTaskFromMap(taskId string) *model.Task {
	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()
	task := hc.status.Tasks[taskId]

	return task
}

func (hc *HealthCheck) getTaskOnline(id string) bool {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	return task.Online
}

func (hc *HealthCheck) setTaskOnline(id string, value bool) {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	task.Online = value
}

func (hc *HealthCheck) getTaskSuccessChecks(id string) int {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	return task.SuccessChecks
}

func (hc *HealthCheck) setTaskSuccessChecks(id string, value int) {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	task.SuccessChecks = value
}

func (hc *HealthCheck) getTaskFailureChecks(id string) int {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	return task.FailureChecks
}

func (hc *HealthCheck) setTaskFailureChecks(id string, value int) {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	task.FailureChecks = value
}

func (hc *HealthCheck) getTaskRestartTime(id string) int64 {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	return task.RestartTime
}

func (hc *HealthCheck) setTaskRestartTime(id string, value int64) {
	task := hc.getTask(id)

	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()

	task.RestartTime = value
}

func (hc *HealthCheck) setTaskInMap(id string, task *model.Task) {
	hc.status.Mx.Lock()
	defer hc.status.Mx.Unlock()
	hc.status.Tasks[id] = task
}

func (hc *HealthCheck) StartTask(function *model.Job) {
	counter := 0
	//task := hc.getTask(function.Id)
	for {
		counter++
		active := false
		if function.DependentJob != "" {
			if hc.getTaskOnline(function.DependentJob) {
				active = true
			}
		} else {
			active = true
		}

		if active {
			if hc.check(function) {
				hc.exporter.SetCounter(function.Id)
				if hc.getTaskOnline(function.Id) {
					log.Info(fmt.Sprintf("%s: Task status updated (is online?): %t",
						function.Id, hc.getTask(function.Id).Online))
				}

				hc.setTaskOnline(function.Id, true)
				hc.setTaskSuccessChecks(function.Id, hc.getTaskSuccessChecks(function.Id)+1)
				hc.setTaskFailureChecks(function.Id, 0)

				log.Debug(fmt.Sprintf("%s: Task status updated (is online?): %t",
					function.Id, hc.getTask(function.Id).Online))
			} else {
				hc.exporter.AddCounter(function.Id, function.Timeout)

				hc.setTaskOnline(function.Id, false)
				hc.setTaskFailureChecks(function.Id, hc.getTaskFailureChecks(function.Id)+1)
				log.Info(fmt.Sprintf("%s: Task status updated (is online?): %t, count: %d",
					function.Id, hc.getTask(function.Id).Online, hc.getTask(function.Id).FailureChecks))

				if function.WatchDogAction.Enabled &&
					hc.getTaskFailureChecks(function.Id) >= function.WatchDogAction.FailureThreshold &&
					(time.Now().Unix()-hc.getTaskRestartTime(function.Id)) > function.WatchDogAction.AwaitAfterRestart {

					log.Info(fmt.Sprintf("Task %s is sent to watchdog", function.Id))
					hc.watchDog.Start(function.WatchDogAction.Actions)

					// for y := 0; y < len(function.WatchDog.Deployments); y++ {
					// 	err := hc.watchDog.DeletePod(function.WatchDog.Deployments[y], function.WatchDog.Namespace)
					// 	if err != nil {
					// 		log.Error(fmt.Sprintf("Delete pod error: %s", err.Error()))
					// 	}
					// }

					hc.setTaskFailureChecks(function.Id, 0)
					hc.setTaskRestartTime(function.Id, time.Now().Unix())
				}
			}
		}

		duration := time.Duration(function.Timeout) * time.Second
		time.Sleep(duration)
	}
}

func (hc *HealthCheck) InitTask(function *model.Job) {
	task := hc.getTask(function.Id)
	log.Info(fmt.Sprintf("%s: Started task", task.Id))

	// if function.Location.Type == "kubernetes" {
	// 	podIps, err := hc.watchDog.GetPodIp(function.Location.Deployment, function.Location.Namespace)
	// 	if err != nil {
	// 		log.Error(fmt.Sprintf("%s: error wss last message exceeded timeout", function.Id))
	// 		return
	// 	}

	// 	urls := make([]string, 0)
	// 	for _, u := range function.Urls {
	// 		base, _ := url.Parse(u)

	// 		for _, ip := range podIps {
	// 			base.Host = fmt.Sprintf("%s:%s", ip, function.Location.Port)
	// 			base.Scheme = "http"
	// 			newurl := base.String()
	// 			//newurl := fmt.Sprintf("%s://%s:%s%s", "http", ip, function.Location.Port, base.Path)
	// 			urls = append(urls, newurl)
	// 		}
	// 	}

	// 	fmt.Println(urls)
	// } else {

	// }
}

func (hc *HealthCheck) check(function *model.Job) bool {
	switch function.Type {
	case "http_get":
		return hc.checkHttpGet(function)
	case "http_post":
		return hc.checkHttpPost(function)
	case "websocket":
		return hc.checkWs(function)
	case "memory":
		return hc.checkMemory(function)
	}

	return false
}

func (hc *HealthCheck) checkMemory(function *model.Job) bool {
	podsMemory, err := hc.cluster.GetPodMemory(function.Label, function.Namespace)
	if err != nil {
		return false
	}

	for i := 0; i < len(podsMemory); i++ {
		if podsMemory[i] > function.Limit {
			log.Error(fmt.Sprintf("Memory usage: %d higher than expected: %d", podsMemory[i], function.Limit))
			return false
		}
	}

	return true
}

func (hc *HealthCheck) checkWs(function *model.Job) bool {
	for _, u := range function.Urls {
		difference := hc.wsClient.TimeDifferenceWithLastMessage(function.Id, u)

		if difference > function.Timeout {
			log.Error(fmt.Sprintf("%s: error wss last message exceeded timeout", function.Id))
			return false
		}
	}

	return true
}

func (hc *HealthCheck) getHttpClient(function *model.Job) *http.Client {
	if function.AuthEnabled {
		return hc.authClient.GetClient()
	} else {
		return hc.httpClient
	}
}

func (hc *HealthCheck) checkHttpGet(function *model.Job) bool {
	start := time.Now()

	for _, u := range function.Urls {
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return false
		}

		if function.ResponseTimeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(function.ResponseTimeout)*time.Second)
			req = req.WithContext(ctx)
			defer cancel()
		}

		resp, err := hc.getHttpClient(function).Do(req)
		if err != nil {
			log.Error(fmt.Sprintf("Error http get request: %s", err.Error()))
			return false
		}
		if resp == nil || resp.StatusCode != 200 {
			log.Error(fmt.Sprintf("%s: Empty http get result or invalid response code", function.Id))
			return false
		}
		defer resp.Body.Close()
	}

	hc.exporter.SetGauge(function.Id, float64(time.Since(start).Milliseconds()))
	log.Info(fmt.Sprintf("%s %s", function.Id, time.Since(start)))

	return true
}

func (hc *HealthCheck) checkHttpPost(function *model.Job) bool {
	for _, u := range function.Urls {
		req, err := http.NewRequest("POST", u, strings.NewReader(function.Body))
		if err != nil {
			return false
		}

		req.Header.Add("accept", "*/*")
		req.Header.Add("Content-Type", "application/json")

		if function.ResponseTimeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(function.ResponseTimeout)*time.Second)
			req = req.WithContext(ctx)
			defer cancel()
		}

		resp, err := hc.getHttpClient(function).Do(req)
		if err != nil {
			log.Error(fmt.Sprintf("Error http post request: %s", err.Error()))
			return false
		}
		if resp == nil || resp.StatusCode != 200 {
			log.Error(fmt.Sprintf("Empty http post result or invalid response code: %d", resp.StatusCode))
			return false
		}
		defer resp.Body.Close()
	}

	return true
}

func (hc *HealthCheck) Status() (*model.Status, error) {
	return hc.status, nil
}

func (hc *HealthCheck) Ready() error {
	return hc.cluster.Test()
}
