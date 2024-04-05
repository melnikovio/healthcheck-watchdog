package model

type Connection struct {
	JodId string
	Url   string
}

type AuthRequest struct {
	AccessToken string `json:"accessToken"`
}

func NewConnection(jobId string, url string) Connection {
	return Connection{JodId: jobId, Url: url}
}
