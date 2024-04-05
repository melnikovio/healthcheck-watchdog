package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/healthcheck-watchdog/cmd/api"
	"github.com/healthcheck-watchdog/cmd/configuration"
	"github.com/healthcheck-watchdog/cmd/exporter"
	"github.com/healthcheck-watchdog/cmd/manager"
	"github.com/healthcheck-watchdog/cmd/watchdog"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

func main() {
	// initialize configuration. panic on error
	config := configuration.NewConfiguration()

	// initialize metrics exporter. panic on error
	exporter := exporter.NewExporter(config)

	// initialize watchdog module
	watchdog := watchdog.NewWatchDog(config)

	// initialize task manager
	manager := manager.NewManager(exporter, watchdog, config)

	// initialize api router
	router := api.NewRouter(manager)

	// start metrics server
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET"},
	})
	server := &http.Server{
		Addr:         ":8080",
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
