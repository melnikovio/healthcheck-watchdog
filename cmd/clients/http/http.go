package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/healthcheck-watchdog/cmd/authentication"
	"github.com/healthcheck-watchdog/cmd/common"
	"github.com/healthcheck-watchdog/cmd/model"

	log "github.com/sirupsen/logrus"
)

type HttpClient struct {
	authClient *authentication.AuthClient
	httpClient *http.Client
	config *model.Config
}

func NewHttpClient(authClient *authentication.AuthClient, config *model.Config) *HttpClient {
	return &HttpClient{
		authClient: authClient,
		httpClient: &http.Client{},
		config: config,
	}
}

func (hc *HttpClient) Execute(job *model.Job, channel chan *model.TaskResult) {
	var method string
	switch job.Type {
	case "http_get":
		method = "GET"
	case "http_post":
		method = "POST"
	}

	for _, u := range job.Urls {
		result := hc.request(job, method, u)
		channel <- result
	}
}

func (hc *HttpClient) request(job *model.Job, method string, url string) *model.TaskResult {
	var requestResult bool
	
	start := time.Now()
	
	var body *strings.Reader
	if job.Body != "" {
		body = strings.NewReader(job.Body)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		requestResult = false
	} else {
		requestResult = hc.performRequest(req, job)
	}

	result := &model.TaskResult{
		Id: job.Id,
		Http: model.Http{
			Url: url,
		},
		Duration: int64(time.Since(start) / time.Millisecond),
		Result: requestResult,
	}

	log.Info(fmt.Sprintf("%s:%s %s", job.Id, url, time.Since(start)))

	return result
}

func (hc *HttpClient) getHttpClient(job *model.Job) *http.Client {
	if job.AuthEnabled {
		return hc.authClient.GetClient()
	} else {
		return hc.httpClient
	}
}

func (hc *HttpClient) performRequest(req *http.Request, job *model.Job) bool {
	req.Header.Add("accept", "*/*")
	req.Header.Add(common.HeaderContentType, common.ContentTypeJson)

	if job.ResponseTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(job.ResponseTimeout)*time.Second)
		req = req.WithContext(ctx)
		defer func() {
			cancel()
		}()
	}

	resp, err := hc.getHttpClient(job).Do(req)
	if resp != nil {
		defer cleanup(resp)
	}
	if err != nil {
		log.Error(fmt.Sprintf("Error http %s request on url %s: %s", req.Method, req.RequestURI, err.Error()))
		return false
	}
	if resp == nil {
		log.Error(fmt.Sprintf("Empty http %s result on url %s", req.Method, req.RequestURI))
		return false
	}
	if resp.StatusCode != 200 {
		log.Error(fmt.Sprintf("Invalid response code %d on url %s", resp.StatusCode, req.RequestURI))
		return false
	}

	return true
}

func cleanup(resp *http.Response) {
	defer resp.Body.Close()
	if resp.Body != nil {
		_, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			log.Error(fmt.Sprintf("Error while read body: %s", err.Error()))
		}
	}
}