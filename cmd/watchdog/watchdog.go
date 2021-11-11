package watchdog

import (
	"fmt"

	"github.com/healthcheck-exporter/cmd/cluster"
	"github.com/healthcheck-exporter/cmd/model"
	"github.com/healthcheck-exporter/cmd/redis"
	log "github.com/sirupsen/logrus"
)

type WatchDog struct {
	cluster *cluster.Cluster
	redis   *redis.Redis
	config  *model.Config
}

func NewWatchDog(config *model.Config, cl *cluster.Cluster) *WatchDog {
	wd := WatchDog{
		cluster: cl,
		redis:   redis.NewRedis(),
		config:  config,
	}

	return &wd
}

func (ws *WatchDog) Start(tasks []string) {
	log.Info(fmt.Sprintf("Started watchdog actions: %v", tasks))

	for i := 0; i < len(tasks); i++ {

		for y := 0; y < len(ws.config.WatchDog.Actions); y++ {
			if ws.config.WatchDog.Actions[y].Id == tasks[i] {
				switch ws.config.WatchDog.Actions[y].Type {
				case "redis":
					ws.redis.Execute(ws.config.WatchDog.Actions[y].ConnectionString, ws.config.WatchDog.Actions[y].Cmd)
				case "deployment_scale_down":
					ws.cluster.ScaleDown(ws.config.WatchDog.Actions[y].Items, ws.config.WatchDog.Namespace)
				case "deployment_scale_up":
					ws.cluster.ScaleUp(ws.config.WatchDog.Actions[y].Items, ws.config.WatchDog.Namespace)
				}
			}
		}

	}
}
