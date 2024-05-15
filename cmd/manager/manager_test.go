package manager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/healthcheck-watchdog/cmd/manager"
	"github.com/healthcheck-watchdog/cmd/model"
)

func TestJobMap_CreateOrUpdate(t *testing.T) {
	manager := &manager.Manager{
		Jobs: make(map[string]*model.TaskStatus),
	}

	// Создание новой задачи
	manager.CreateOrUpdateTask("task1", nil, func(task *model.TaskStatus) {
		task.Id = "Task 1"
		task.Running = true
	})

	// Проверка, что задача создана
	assert.NotNil(t, manager.GetTask("task1"))

	// Обновление существующей задачи
	manager.CreateOrUpdateTask("task1", nil, func(task *model.TaskStatus) {
		task.Running = false
	})

	// Проверка, что задача обновлена
	assert.Equal(t, false, manager.GetTask("task1").Running)
}

func TestJobMap_CreateOrUpdate_NewTask(t *testing.T) {
	manager := &manager.Manager{
		Jobs: make(map[string]*model.TaskStatus),
	}

	// Создание новой задачи с использованием функции обновления
	manager.CreateOrUpdateTask("task1", nil, func(task *model.TaskStatus) {
		task.Id = "Task 1"
		task.Running = true
	})

	// Проверка, что задача создана
	assert.NotNil(t, manager.GetTask("task1"))
	assert.Equal(t, "Task 1", manager.GetTask("task1").Id)
	assert.Equal(t, true, manager.GetTask("task1").Running)
}

func TestJobMap_CreateOrUpdate_UpdateTask(t *testing.T) {
	manager := &manager.Manager{
		Jobs: map[string]*model.TaskStatus{
			"task1": {Id: "Task 1", Running: true},
		},
	}

	// Обновление существующей задачи с использованием функции обновления
	manager.CreateOrUpdateTask("task1", nil, func(task *model.TaskStatus) {
		task.Running = false
	})

	// Проверка, что задача обновлена
	assert.Equal(t, false, manager.GetTask("task1").Running)
}
