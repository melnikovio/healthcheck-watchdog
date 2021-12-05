package healthcheck

import (
	"fmt"
	"sync"
	"time"

	"github.com/healthcheck-watchdog/cmd/exporter"
	"github.com/sacOO7/gowebsocket"
	log "github.com/sirupsen/logrus"
)

type WsClient struct {
	mx          sync.Mutex
	connections map[string]*WsConnection
	prometheus  *exporter.Exporter
}

func (wc *WsClient) getConnection(key string) *WsConnection {
	wc.mx.Lock()
	defer wc.mx.Unlock()

	return wc.connections[key]
}

func (wc *WsClient) setConnection(key string, value *WsConnection) {
	wc.mx.Lock()
	defer wc.mx.Unlock()

	wc.connections[key] = value
}

type WsConnection struct {
	mx   sync.Mutex
	urls map[string]*Url
}

type Url struct {
	url  string
	time int64
}

func (wc *WsConnection) getUrl(key string) *Url {
	wc.mx.Lock()
	defer wc.mx.Unlock()

	return wc.urls[key]
}

func (wc *WsConnection) setUrl(key string, value *Url) {
	wc.mx.Lock()
	defer wc.mx.Unlock()

	wc.urls[key] = value
}

// func (wc *WsConnection) getUrlString(key string) string {
// 	wc.mx.Lock()
// 	defer wc.mx.Unlock()

// 	return wc.urls[key].url
// }

// func (wc *WsConnection) getUrlTime(key string) int64 {
// 	wc.mx.Lock()
// 	defer wc.mx.Unlock()

// 	return wc.urls[key].time
// }

func (wc *WsConnection) setUrlTime(key string, value int64) {
	wc.mx.Lock()
	defer wc.mx.Unlock()

	url := &Url{
		url:  key,
		time: value,
	}

	wc.urls[key] = url
}

func NewWsClient(prometheus *exporter.Exporter) *WsClient {
	connection := make(map[string]*WsConnection)
	wc := WsClient{
		connections: connection,
		prometheus:  prometheus,
	}

	return &wc
}

func (ws *WsClient) getWsConnection(jobId string) *WsConnection {
	connection := ws.getConnection(jobId)
	if connection == nil {
		connection = &WsConnection{
			urls: make(map[string]*Url),
		}
		ws.setConnection(jobId, connection)
	}

	return connection
}

func (ws *WsClient) getWsUrl(jobId string, urlAddress string) *Url {
	connection := ws.getWsConnection(jobId)
	url := connection.getUrl(urlAddress)
	if url == nil {
		url = &Url{
			url:  urlAddress,
			time: time.Now().Unix(),
		}
		connection.setUrl(urlAddress, url)
		ws.addWsUrl(url.url, connection, jobId)
	}

	return url
}

func (ws *WsClient) deleteWsUrl(jobId string, urlAddress string) {
	connection := ws.getWsConnection(jobId)
	if connection.urls[urlAddress] != nil {
		delete(connection.urls, urlAddress)
	}
}

func (wsClient *WsClient) addWsUrl(url string, connection *WsConnection, jobId string) {
	log.Info(fmt.Sprintf("Registering url: %s", url))
	socket := gowebsocket.New(url)

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		socket.Close()
		wsClient.deleteWsUrl(jobId, url)
		//wsClient.getWsUrl(jobId, url)
		log.Fatal("Received connect error - ", err)
	}

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println(jobId + ": Received message - " + message)
		wsClient.prometheus.IncCounter(jobId)
		connection.setUrlTime(url, time.Now().Unix())
	}

	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received ping - " + data)
	}

	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong - " + data)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		socket.Close()
		wsClient.deleteWsUrl(jobId, url)
		//wsClient.getWsUrl(jobId, url)
	}

	socket.Connect()
}

func (ws *WsClient) TimeLastMessage(jobId string, url string) int64 {
	return ws.getWsUrl(jobId, url).time
}

func (ws *WsClient) TimeDifferenceWithLastMessage(jobId string, url string) int64 {
	return time.Now().Unix() - ws.getWsUrl(jobId, url).time
}
