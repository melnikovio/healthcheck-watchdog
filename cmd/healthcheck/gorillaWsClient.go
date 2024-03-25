package healthcheck

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/healthcheck-watchdog/cmd/authentication"
	"github.com/healthcheck-watchdog/cmd/exporter"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type GorillaWsClient struct {
	mx          sync.Mutex
	connections map[string]*WsConnection
	prometheus  *exporter.Exporter
	authClient  *authentication.AuthClient
}

func NewGorillaWsClient(prometheus *exporter.Exporter, authClient *authentication.AuthClient) *GorillaWsClient {
	connection := make(map[string]*WsConnection)
	wc := GorillaWsClient{
		connections: connection,
		prometheus:  prometheus,
		authClient:  authClient,
	}

	return &wc
}

func (wc *GorillaWsClient) getUrl(jobId string, urlAddress string, responseTimeout int) *Url {
	connection := wc.getConnection(jobId)
	url := connection.getUrl(urlAddress)
	if url == nil {
		log.Info(fmt.Sprintf("%s. Creating url: %s", jobId, urlAddress))

		url = &Url{
			url:  urlAddress,
			time: time.Now().Unix(),
		}
		connection.setUrl(urlAddress, url)
		wc.addUrl(jobId, url.url, responseTimeout)
	}

	return url
}

func (wc *GorillaWsClient) deleteUrl(jobId string, urlAddress string) {
	connection := wc.getConnection(jobId)
	if connection.urls[urlAddress] != nil {
		delete(connection.urls, urlAddress)
	}
}

func (wc *GorillaWsClient) getConnection(key string) *WsConnection {
	wc.mx.Lock()
	defer wc.mx.Unlock()

	if wc.connections[key] == nil {
		wc.connections[key] = &WsConnection{
			urls: make(map[string]*Url),
		}
	}

	return wc.connections[key]
}

type AuthRequest struct {
	AccessToken string `json:"accessToken"`
}

func (wc *GorillaWsClient) addUrl(jobId string, url string, responseTimeout int) {
	log.Info(fmt.Sprintf("%s. Registering url: %s", jobId, url))
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Error(fmt.Sprintf("%s. Received connect error: %s", jobId, err.Error()))
	}

	//todo depending on config
	token := wc.authClient.GetToken().AccessToken
	auth := AuthRequest{AccessToken: token}
	jsonData, _ := json.Marshal(auth)

	err = c.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Error(fmt.Sprintf("%s. Received connect error: %s", jobId, err.Error()))
	}

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Error(fmt.Sprintf("%s. Received ws (%s) error: %s", jobId, url, err.Error()))
				err := c.Close()
				if err != nil {
					log.Error(fmt.Sprintf("%s. Received ws (%s) error on close: %s", jobId, url, err.Error()))
				}
				wc.deleteUrl(jobId, url)
				return
			}
			log.Info(fmt.Sprintf("%s. Received message: %s", jobId, message))

			var params string
			var data []Object
			if err := json.Unmarshal(message, &data); err != nil {
				log.Error(fmt.Sprintf("%s. failed to unmarshal: %s", jobId, message))
			} else {
				//todo config
				if len(data) > 0 && data[0]["uid"] != nil {
					params = data[0]["uid"].(string)
				}
			}

			wc.prometheus.IncCounter(jobId, params)
			wc.getConnection(jobId).setUrlTime(url, time.Now().Unix())
		}
	}()

	if responseTimeout != 0 {
		for {
			difference := wc.TimeDifferenceWithLastMessage(jobId, url, responseTimeout)

			if difference > int64(responseTimeout) {
				log.Error(fmt.Sprintf("%s: error wss reached response timeout. Closing connection", jobId))
				err := c.Close()
				if err != nil {
					log.Error(fmt.Sprintf("%s. Received ws (%s) error on close: %s", jobId, url, err.Error()))
				}
				wc.deleteUrl(jobId, url)
				wc.addUrl(jobId, url, responseTimeout)
			}

			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}

func (wc *GorillaWsClient) TimeDifferenceWithLastMessage(jobId string, url string, responseTimeout int) int64 {
	return time.Now().Unix() - wc.getUrl(jobId, url, responseTimeout).time
}

// todo
type Object map[string]interface{}
