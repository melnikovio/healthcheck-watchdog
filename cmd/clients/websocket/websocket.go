package clients

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/healthcheck-watchdog/cmd/authentication"
	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
)

type WsClient struct {
	authClient  *authentication.AuthClient
	config      *model.Config
	connections *Connections
}

func NewWsClient(authClient *authentication.AuthClient, config *model.Config) *WsClient {
	return &WsClient{
		authClient:  authClient,
		config:      config,
		connections: NewConnections(),
	}
}

// SafeMap for connections pool
type Connections struct {
	mu          sync.RWMutex
	connections map[model.Connection]*websocket.Conn
}

func NewConnections() *Connections {
	return &Connections{
		connections: make(map[model.Connection]*websocket.Conn),
	}
}

// Get connection
func (m *Connections) Get(key model.Connection) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.connections[key]
	return value, ok
}

// Set connection
func (m *Connections) Set(key model.Connection, value *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[key] = value
}

// Delete connection
func (m *Connections) Delete(key model.Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, key)
}

// Execute job
func (wc *WsClient) Execute(job *model.Job, channel chan *model.TaskResult) {
	for _, u := range job.Urls {
		runningJob := model.CreateRunningJob(job, u)
		connection := model.NewConnection(job.Id, u)
		_, ok := wc.connections.Get(connection)
		if !ok {
			wc.connections.Set(connection, &websocket.Conn{})
			wc.connect(runningJob, channel)
		}
	}
}

// Connect to ws
func (wc *WsClient) connect(job *model.RunningJob, channel chan *model.TaskResult) {
	log.Info(fmt.Sprintf("%s. Registering websocket url: %s", job.Id, job.Url))

	// Connect to WS
	c, _, err := websocket.DefaultDialer.Dial(job.Url, nil)
	if err != nil {
		log.Error(fmt.Sprintf("%s. Received connect error: %s", job.Id, err.Error()))

		wc.connections.Delete(model.NewConnection(job.Id, job.Url))
		result := &model.TaskResult{
			Id:      job.Id,
			Result:  false,
			Running: false,
		}
		channel <- result

		return
	}

	// Send authentication message
	if job.Auth.Enabled {
		auth := model.AuthRequest{AccessToken: wc.authClient.GetToken(job.Auth.Client).AccessToken}
		authMessage, _ := json.Marshal(auth)

		err = c.WriteMessage(websocket.TextMessage, authMessage)
		if err != nil {
			log.Error(fmt.Sprintf("%s. Received authentication error: %s", job.Id, err.Error()))
		}
	}

	// // Set deadline for messages
	// if err := c.SetReadDeadline(
	// 	time.Now().Add(time.Duration(job.ResponseTimeout) * time.Second)); err != nil {
	// 	log.Error(fmt.Sprintf("%s. Set read deadline error: %s", job.Id, err.Error()))
	// }

	// Receive messages
	go func() {
		start := time.Now()
		for {
			// Set read timer
			if err := c.SetReadDeadline(
				time.Now().Add(time.Duration(job.ResponseTimeout) * time.Second)); err != nil {
				log.Error(fmt.Sprintf("%s. Set read deadline error: %s", job.Id, err.Error()))
			}

			// Read messages
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Error(fmt.Sprintf("%s. Received ws error (%s): %s", job.Id, job.Url, err.Error()))
				err := c.Close()
				if err != nil {
					log.Error(fmt.Sprintf("%s. Received ws error (%s) on close: %s", job.Id, job.Url, err.Error()))
				}

				log.Info(fmt.Sprintf("%s. Websocket closed", job.Id))

				wc.connections.Delete(model.NewConnection(job.Id, job.Url))

				result := &model.TaskResult{
					Id:       job.Id,
					Result:   false,
					Running:  false,
					Duration: time.Since(start).Milliseconds(),
				}
				channel <- result

				return
			}

			log.Info(fmt.Sprintf("%s. Received message: %s", job.Id, message))

			result := &model.TaskResult{
				Id:       job.Id,
				Result:   true,
				Running:  true,
				Duration: time.Since(start).Milliseconds(),
			}

			var data []map[string]interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				log.Error(fmt.Sprintf("%s. failed to unmarshal: %s", job.Id, message))
			} else {
				if len(data) > 0 {
					result.Parameters = data[0]
				}
			}

			channel <- result

			// Reset messages timer
			start = time.Now()
		}
	}()
}
