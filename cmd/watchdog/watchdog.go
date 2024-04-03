package watchdog

import (
	"fmt"

	"github.com/healthcheck-watchdog/cmd/cluster"
	"github.com/healthcheck-watchdog/cmd/model"
	"github.com/healthcheck-watchdog/cmd/redis"
	log "github.com/sirupsen/logrus"
)

type WatchDog struct {
	cluster *cluster.Cluster
	redis   *redis.Redis
	config  *model.Config
	Channel chan *model.TaskResult
}

func NewWatchDog(config *model.Config) *WatchDog {
	if config.WatchDog.Namespace == "" && 
		len(config.WatchDog.Actions) == 0 {
			log.Info("Missing watchdog configuration. Watchdog configuration ignored.")
			return nil
	}

	wd := WatchDog{
		// cluster: cl,
		redis:   redis.NewRedis(),
		Channel: make(chan *model.TaskResult),
		config:  config,
	}

	// go wd.receiver()
	// Channel for task results
	go wd.resultProcessor(wd.Channel)

	return &wd
}


// func (wd *WatchDog) Message(message string) {
// 	select {
// 	case wd.channel <- message:
// 		log.Trace("")
// 	default:
// 		log.Error("")
// 	}
// }

// func (wd *WatchDog) receiver() {
// 	for {
// 		// Получаем сообщение из канала
// 		msg, ok := <- wd.channel
// 		if !ok {
// 			log.Error("Channel closed, exiting receiver")
// 			return
// 		}
// 		fmt.Println("Received message:", msg)
// 	}
// }


// Process results from tasks
func (ws *WatchDog) resultProcessor(resultChan <-chan *model.TaskResult) {
	for result := range resultChan {
		log.Info(fmt.Sprintf("Watchdog: Processed result %v", result))
	}
}

func (ws *WatchDog) Execute(tasks []string) {
	log.Info(fmt.Sprintf("Started watchdog actions: %v", tasks))

	for i := range tasks {
		for y := range ws.config.WatchDog.Actions {
			var err error
			if ws.config.WatchDog.Actions[y].Id == tasks[i] {
				switch ws.config.WatchDog.Actions[y].Type {
				case "redis":
					err = ws.redis.Execute(ws.config.WatchDog.Actions[y].ConnectionString, ws.config.WatchDog.Actions[y].Cmd)
				case "deployment_scale_down":
					err = ws.cluster.ScaleDown(ws.config.WatchDog.Actions[y].Items, ws.config.WatchDog.Namespace)
				case "deployment_scale_up":
					err = ws.cluster.ScaleUp(ws.config.WatchDog.Actions[y].Items, ws.config.WatchDog.Namespace)
				}
			}

			if err != nil {
				log.Error(fmt.Sprintf("Error in task %s: %s", ws.config.WatchDog.Actions[y].Id, err.Error()))
			}
		}
	}
}
