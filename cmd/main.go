package main

import (
	"fmt"
	"net/http"

	"github.com/healthcheck-exporter/cmd/api"
	"github.com/healthcheck-exporter/cmd/authentication"
	"github.com/healthcheck-exporter/cmd/cluster"
	"github.com/healthcheck-exporter/cmd/common"
	"github.com/healthcheck-exporter/cmd/configuration"
	"github.com/healthcheck-exporter/cmd/exporter"
	"github.com/healthcheck-exporter/cmd/healthcheck"
	"github.com/healthcheck-exporter/cmd/watchdog"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Println(common.Logo)

	// initialize configuration. panic if error
	config := configuration.NewConfiguration()

	// initialize auth client. panic if error
	auth := authentication.NewAuthClient(config)

	// initialize metrics exporter. panic if error
	exporter := exporter.NewExporter(config)

	// initialize cluster client. panic if error
	cluster := cluster.NewCluster()

	// initialize watchdog app. panic if error
	watchdog := watchdog.NewWatchDog(config, cluster)

	// initialize healthcheck. panic if error
	healthcheck := healthcheck.NewHealthCheck(config, auth, exporter, watchdog, cluster)

	// initialize api router
	router := api.NewRouter(healthcheck)

	// enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET"},
	})

	log.Info(fmt.Sprintf("%v", http.ListenAndServe(":2112",
		corsHandler.Handler(router)).Error()))
}
