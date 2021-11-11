package main

import (
	"fmt"
	"net/http"

	"github.com/healthcheck-watchdog/cmd/api"
	"github.com/healthcheck-watchdog/cmd/authentication"
	"github.com/healthcheck-watchdog/cmd/cluster"
	"github.com/healthcheck-watchdog/cmd/common"
	"github.com/healthcheck-watchdog/cmd/configuration"
	"github.com/healthcheck-watchdog/cmd/exporter"
	"github.com/healthcheck-watchdog/cmd/healthcheck"
	"github.com/healthcheck-watchdog/cmd/watchdog"
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
