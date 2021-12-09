package healthcheck

import (
	"fmt"
	"sync"
	"time"

	"github.com/healthcheck-watchdog/cmd/exporter"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type GorillaWsClient struct {
	mx          sync.Mutex
	connections map[string]*WsConnection
	prometheus  *exporter.Exporter
}

func NewGorillaWsClient(prometheus *exporter.Exporter) *GorillaWsClient {
	connection := make(map[string]*WsConnection)
	wc := GorillaWsClient{
		connections: connection,
		prometheus:  prometheus,
	}

	return &wc
}

func (wc *GorillaWsClient) getUrl(jobId string, urlAddress string) *Url {
	connection := wc.getConnection(jobId)
	url := connection.getUrl(urlAddress)
	if url == nil {
		log.Info(fmt.Sprintf("%s. Creating url: %s", jobId, urlAddress))

		url = &Url{
			url:  urlAddress,
			time: time.Now().Unix(),
		}
		connection.setUrl(urlAddress, url)
		wc.addUrl(jobId, url.url)
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

func (wc *GorillaWsClient) addUrl(jobId string, url string) {
	log.Info(fmt.Sprintf("%s. Registering url: %s", jobId, url))
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Error(fmt.Sprintf("%s. Received connect error: %s", jobId, err.Error()))
	}

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				c.Close()
				log.Error(fmt.Sprintf("%s. Received ws (%s) error: %s", jobId, url, err.Error()))
				wc.deleteUrl(jobId, url)
				return
			}
			log.Trace(fmt.Sprintf("%s. Received message: %s", jobId, message))
			wc.prometheus.IncCounter(jobId)
			wc.getConnection(jobId).setUrlTime(url, time.Now().Unix())
		}
	}()
}

func (wc *GorillaWsClient) TimeDifferenceWithLastMessage(jobId string, url string) int64 {
	return time.Now().Unix() - wc.getUrl(jobId, url).time
}
