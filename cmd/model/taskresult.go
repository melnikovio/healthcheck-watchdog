package model

type TaskResult struct {
	Id string
	Duration int64
	Result bool
	Http Http
	Websocket Websocket
	Kubernetes Kubernetes
	Parameters map[string]interface{}
}

type Http struct {
	Url string
}

type Websocket struct {
	Url string
}

type Kubernetes struct {
	
}
