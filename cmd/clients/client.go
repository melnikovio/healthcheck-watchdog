package clients

import (
	"github.com/healthcheck-watchdog/cmd/model"
)

// Interface for clients
type Executor interface {
	Execute(job *model.Job, channel chan *model.TaskResult)
}
