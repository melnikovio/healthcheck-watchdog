package model

import "sync"

//swagger:model
type Status struct {
	Mx sync.Mutex
	// required: true
	Tasks map[string]*Task `json:"tasks,omitempty"`
}

//swagger:model
type Task struct {
	// required: true
	Id string `json:"id,omitempty"`
	// required: true
	Online bool `json:"online,omitempty"`
	// required: true
	SuccessChecks int `json:"success_checks,omitempty"`
	// required: true
	FailureChecks int `json:"failure_checks,omitempty"`
	// required: true
	RestartTime int64 `json:"restartTime,omitempty"`
}
