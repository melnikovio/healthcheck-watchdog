package model

type Action struct {
	// required: true
	Id string `json:"id,omitempty"`
	// required: true
	Type string `json:"type,omitempty"`
	// required: true
	ConnectionString string `json:"connectionstring,omitempty"`
	// required: true
	Cmd string `json:"cmd,omitempty"`
	// required: true
	Items []string `json:"items,omitempty"`
}
