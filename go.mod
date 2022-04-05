module github.com/healthcheck-watchdog

go 1.16

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/client_model v0.2.0
	github.com/rs/cors v1.8.2
	github.com/sacOO7/gowebsocket v0.0.0-20210515122958-9396f1a71e23
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	k8s.io/metrics v0.23.5
)
