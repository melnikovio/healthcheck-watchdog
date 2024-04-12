package model

type Job struct {
	// required: true
	Id string `json:"id,omitempty"`
	// required: true
	Description string `json:"desc,omitempty"`
	// required: true
	Type string `json:"type,omitempty"`
	// required: true
	Label string `json:"label,omitempty"`
	// required: true
	Namespace string `json:"namespace,omitempty"`
	// required: true
	Limit int64 `json:"limit,omitempty"`
	// required: true
	Urls []string `json:"urls,omitempty"`
	// required: true
	Body string `json:"body,omitempty"`
	// required: true
	// deprecated: true
	AuthEnabled bool `json:"auth_enabled,omitempty"`
	// required: true
	Auth Auth `json:"auth,omitempty"`
	// required: true
	Timeout int64 `json:"timeout,omitempty"`
	// required: true
	ResponseTimeout int `json:"responseTimeout,omitempty"`
	// required: true
	DependentJob string `json:"dependentJob,omitempty"`
	// required: true
	Location Location `json:"location,omitempty"`
	// required: true
	WatchDogAction WatchDogAction `json:"watchdog_action,omitempty"`

	// required: true
	MetricName string `json:"metricName,omitempty"`

	MetricLabels map[string]string `json:"metricLabels,omitempty"`
}

type Auth struct {
	Enabled bool   `json:"enabled,omitempty"`
	Client  string `json:"client,omitempty"`
}
