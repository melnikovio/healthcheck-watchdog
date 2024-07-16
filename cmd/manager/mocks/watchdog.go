package manager_test

import (
	"github.com/healthcheck-watchdog/cmd/model"
)


type MockWatchdog struct {
	channel      chan model.TaskStatus
	Tasks int
}

func NewMockWatchdog(config *model.Config) *MockWatchdog {
	exporter := MockWatchdog{
		channel: make(chan model.TaskStatus, len(config.Jobs)),
	}
	
	go exporter.resultProcessor(exporter.channel)

	return &exporter
}

func (ex *MockWatchdog) GetChannel() chan model.TaskStatus {
	return ex.channel
}

func (ex *MockWatchdog) resultProcessor(resultChan <-chan model.TaskStatus) {
	for range resultChan {
		ex.Tasks++
	}
}