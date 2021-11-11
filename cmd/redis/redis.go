package redis

import (
	"context"
	"fmt"

	rediscli "github.com/go-redis/redis/v8"
	"github.com/healthcheck-watchdog/cmd/common"
	log "github.com/sirupsen/logrus"
)

type Redis struct {
}

func NewRedis() *Redis {
	return &Redis{}
}

func (redis *Redis) connect(cs string) *rediscli.Client {
	r := rediscli.NewClient(&rediscli.Options{
		Addr:     cs,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return r
}

func (redis *Redis) Execute(cs string, cmd string) {
	log.Info(fmt.Sprintf("Started task on redis: %s", cs))

	client := redis.connect(cs)

	var result *rediscli.StatusCmd

	switch cmd {
	case common.RedisFlushAll:
		result = client.FlushAll(context.Background())
	}

	log.Info(fmt.Sprintf("%v", result))
}
