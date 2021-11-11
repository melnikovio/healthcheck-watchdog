package model

type Location struct {
	// required: true
	Type string `json:"type,omitempty"`
	// required: true
	Deployment string `json:"deployment,omitempty"`
	// required: true
	Namespace string `json:"namespace,omitempty"`
	// required: true
	Port string `json:"port,omitempty"`
}
