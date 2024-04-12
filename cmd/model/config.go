package model

//swagger:model
type Config struct {
	// required: false
	// deprecated: true
	Authentication Authentication `json:"authentication,omitempty"`

	// required: false
	AuthenticationClients map[string]Authentication `json:"authenticationClients,omitempty"`

	// required: true
	Jobs []Job `json:"jobs,omitempty"`

	// required: false
	WatchDog WatchDog `json:"watchdog,omitempty"`

	// required: false
	LogLevel string `json:"loglevel,omitempty"`
}

type Authentication struct {
	// required: true
	AuthUrl string `json:"auth_url,omitempty"`
	// required: true
	ClientId string `json:"client_id,omitempty"`
	// required: true
	ClientSecret string `json:"client_secret,omitempty"`
}
