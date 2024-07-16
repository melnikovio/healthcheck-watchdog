package manager_test

import (
	"github.com/healthcheck-watchdog/cmd/model"
)


type MockExporter struct {
	channel      chan model.TaskStatus
	Tasks int
}

func NewMockExporter(config *model.Config) *MockExporter {
	exporter := MockExporter{
		channel: make(chan model.TaskStatus, len(config.Jobs)),
	}
	
	go exporter.resultProcessor(exporter.channel)

	return &exporter
}

func (ex *MockExporter) GetChannel() chan model.TaskStatus {
	return ex.channel
}

func (ex *MockExporter) resultProcessor(resultChan <-chan model.TaskStatus) {
	for range resultChan {
		ex.Tasks++
	}
}