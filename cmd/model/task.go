package model

type TaskStatus struct {
	Id         string
	Running    bool
	LastCall   int64
	Status     bool
	LastResult *TaskResult
	Job        *Job
	Watchdog   bool
}
