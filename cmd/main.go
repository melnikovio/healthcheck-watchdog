package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/healthcheck-watchdog/cmd/api"
	"github.com/healthcheck-watchdog/cmd/authentication"
	"github.com/healthcheck-watchdog/cmd/configuration"
	"github.com/healthcheck-watchdog/cmd/exporter"
	"github.com/healthcheck-watchdog/cmd/healthcheck"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

func main() {
	// initialize configuration. panic on error
	config := configuration.NewConfiguration()

	// initialize auth client. panic on error
	authClient := authentication.NewAuthClient(config)

	// initialize metrics exporter. panic on error
	exporter := exporter.NewExporter(config)

	// initialize cluster client. panic if error
	//cluster := cluster.NewCluster()

	// initialize watchdog app. panic if error
	//watchdog := watchdog.NewWatchDog(cluster, config)

	// initialize healthcheck. panic if error
	healthcheck := healthcheck.NewHealthCheck(config, authClient, exporter, nil, nil)

	// initialize api router
	router := api.NewRouter(healthcheck)

	// enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET"},
	})

	// start metrics server
	server := &http.Server{
		Addr:         ":2112",
		Handler:      corsHandler.Handler(router),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Info(fmt.Sprintf("HTTP server started on http://localhost%s", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		log.Error(fmt.Sprintf("HTTP server error: %s", err.Error()))
		panic(err)
	}
}
