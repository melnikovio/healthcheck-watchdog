package model

type TaskResult struct {
	Id         string
	Duration   int64
	Result     bool
	Running    bool
	Http       Http
	Kubernetes Kubernetes
	Parameters map[string]interface{}
}

type Http struct {
	Url string
}

type Kubernetes struct {
}
