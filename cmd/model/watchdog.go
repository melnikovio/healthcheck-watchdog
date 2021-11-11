package model

type WatchDog struct {
	// required: true
	Namespace string `json:"namespace,omitempty"`
	// required: true
	Actions []Action `json:"actions,omitempty"`
}
