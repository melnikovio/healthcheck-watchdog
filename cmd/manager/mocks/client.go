package manager_test

import (
	"github.com/healthcheck-watchdog/cmd/model"
)

type MockClient struct {
	channel chan model.TaskStatus
}

func NewMockClient(config *model.Config) *MockClient {
	return &MockClient{
		channel: make(chan model.TaskStatus, len(config.Jobs)),
	}
}

func (mc *MockClient) Execute(job *model.Job, channel chan *model.TaskResult) {
	result := &model.TaskResult{
		Id:      job.Id,
		Running: false,
		Result:  true,
	}
	channel <- result
}
