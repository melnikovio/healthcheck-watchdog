package model

type WatchDogAction struct {
	// required: true
	Enabled bool `json:"enabled,omitempty"`
	// required: true
	Actions []string `json:"actions,omitempty"`
	// required: true
	FailureThreshold int `json:"failureThreshold,omitempty"`
	// required: true
	AwaitAfterRestart int64 `json:"awaitAfterRestart,omitempty"`
}
