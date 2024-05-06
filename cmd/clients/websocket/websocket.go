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

// SafeMap представляет потокобезопасную map[string]string
type Connections struct {
	mu          sync.RWMutex
	connections map[model.Connection]*websocket.Conn
}

// NewSafeMap создает новый экземпляр SafeMap
func NewConnections() *Connections {
	return &Connections{
		connections: make(map[model.Connection]*websocket.Conn),
	}
}

// Get возвращает значение по ключу из SafeMap
func (m *Connections) Get(key model.Connection) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.connections[key]
	return value, ok
}

// Set устанавливает значение по ключу в SafeMap
func (m *Connections) Set(key model.Connection, value *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[key] = value
}

// Delete удаляет значение по ключу из SafeMap
func (m *Connections) Delete(key model.Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, key)
}

func (wc *WsClient) Execute(job *model.Job, channel chan *model.TaskResult) {
	for _, u := range job.Urls {
		runningJob := model.CreateRunningJob(job, u)
		connection := model.NewConnection(job.Id, u)
		_, ok := wc.connections.Get(connection)
		if !ok {
			wc.connections.Set(connection, wc.connect(runningJob, channel))
		}
	}
}

func (wc *WsClient) connect(job *model.RunningJob, channel chan *model.TaskResult) *websocket.Conn {
	log.Info(fmt.Sprintf("%s. Registering websocket url: %s", job.Id, job.Url))

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

		return nil
	}

	if job.Auth.Enabled {
		auth := model.AuthRequest{AccessToken: wc.authClient.GetToken(job.Auth.Client).AccessToken}
		authMessage, _ := json.Marshal(auth)

		err = c.WriteMessage(websocket.TextMessage, authMessage)
		if err != nil {
			log.Error(fmt.Sprintf("%s. Received authentication error: %s", job.Id, err.Error()))
		}
	}

	// var resetTimer func()
	// if job.ResponseTimeout != 0 {
	// 	// Функция для перезапуска таймера
	// 	var timer *time.Timer
	// 	resetTimer = func() {
	// 		if timer != nil {
	// 			timer.Stop()
	// 		}
	// 		timer = time.AfterFunc(time.Duration(job.ResponseTimeout)*time.Second, func() {
	// 			log.Error(
	// 				fmt.Sprintf("%s. No messages received in %d seconds. Closing connection.",
	// 					job.Id, job.ResponseTimeout))

	// 			c.SetD
	// 			err := c.Close()
	// 			if err != nil {
	// 				log.Error(fmt.Sprintf("%s. Received ws error (%s) on close: %s", job.Id, job.Url, err.Error()))
	// 			}

	// 			wc.connections.Delete(model.NewConnection(job.Id, job.Url))
	// 			result := &model.TaskResult{
	// 				Id:      job.Id,
	// 				Result:  false,
	// 				Running: false,
	// 			}
	// 			channel <- result
	// 		})
	// 	}
	// }

	c.SetReadDeadline(time.Now().Add(time.Duration(job.ResponseTimeout)*time.Second))

	go func() {
		start := time.Now()
		for {
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

			// if resetTimer != nil && false {
			// 	resetTimer()
			// }

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

			c.SetReadDeadline(time.Now().Add(time.Duration(job.ResponseTimeout)*time.Second))
			
			start = time.Now()
		}
	}()

	return c
}