package manager_test

import (
	"github.com/healthcheck-watchdog/cmd/model"
)

type MockFloodClient struct {
	channel      chan model.TaskStatus
}

func NewMockFloodClient(config *model.Config) *MockFloodClient {
	return &MockFloodClient{
		channel: make(chan model.TaskStatus, len(config.Jobs)),
	}
}

func (mc *MockFloodClient) Execute(job *model.Job, channel chan *model.TaskResult) {
	result := &model.TaskResult{
		Id:      job.Id,
		Running: true,
		Result:   true,
	}

	for i := 0; i < 100; i++ {
		channel <- result
	}

	result.Running = false
	channel <- result
}
