package model

type SchedulerTask struct {
	Id string
	Running bool
	LastStatus bool
	LastCall int64
}