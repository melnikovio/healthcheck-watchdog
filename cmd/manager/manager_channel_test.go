package manager_test

import (
	"testing"
	"time"

	"github.com/healthcheck-watchdog/cmd/clients"
	"github.com/healthcheck-watchdog/cmd/manager"
	mocks "github.com/healthcheck-watchdog/cmd/manager/mocks"
	"github.com/healthcheck-watchdog/cmd/model"
	"github.com/stretchr/testify/assert"
)

var emptyConfig = model.Config{
	Jobs: []model.Job{},
}

var configOneJob = model.Config{
	Jobs: []model.Job{
		{
			Id:      "test",
			Type:    "http_get",
			Timeout: 1,
		},
	},
}

func TestChannels_NewManager(t *testing.T) {
	manager := &manager.Manager{
		Config:    &emptyConfig,
		Executors: make(map[string]clients.Executor),
		Jobs:      make(map[string]*model.TaskStatus),
		Exporter:  mocks.NewMockExporter(&emptyConfig),
	}

	assert.NotEqual(t, manager, nil)
}

func TestChannels_EmptyConfig(t *testing.T) {
	manager := &manager.Manager{
		Config:    &emptyConfig,
		Executors: make(map[string]clients.Executor),
		Jobs:      make(map[string]*model.TaskStatus),
		Exporter:  mocks.NewMockExporter(&emptyConfig),
	}

	manager.Run()

	assert.Equal(t, manager.Exporter.(*mocks.MockExporter).Tasks, 0)
}

func TestChannels_ManagerOneJob(t *testing.T) {
	manager := &manager.Manager{
		Config:    &configOneJob,
		Executors: make(map[string]clients.Executor),
		Jobs:      make(map[string]*model.TaskStatus),
		Exporter:  mocks.NewMockExporter(&configOneJob),
	}

	manager.Executors["http_get"] = mocks.NewMockClient(&configOneJob)

	manager.Run()

	time.Sleep(1 * time.Second)

	assert.Equal(t, manager.Exporter.(*mocks.MockExporter).Tasks, 1)
}

func TestChannels_ManagerOneJobWithFlood(t *testing.T) {
	manager := &manager.Manager{
		Config:    &configOneJob,
		Executors: make(map[string]clients.Executor),
		Jobs:      make(map[string]*model.TaskStatus),
		Exporter:  mocks.NewMockExporter(&configOneJob),
		Watchdog:  mocks.NewMockWatchdog(&configOneJob),
	}

	manager.Executors["http_get"] = mocks.NewMockFloodClient(&configOneJob)

	manager.Run()

	time.Sleep(1 * time.Second)

	assert.Equal(t, manager.Exporter.(*mocks.MockExporter).Tasks, 101)
	assert.Equal(t, manager.Watchdog.(*mocks.MockWatchdog).Tasks, 101)
}
