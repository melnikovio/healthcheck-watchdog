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
}

func NewWatchDog(cl *cluster.Cluster, config *model.Config) *WatchDog {
	wd := WatchDog{
		cluster: cl,
		redis:   redis.NewRedis(),
		config:  config,
	}

	return &wd
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
