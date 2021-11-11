package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

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

	config := configuration.GetConfiguration()
	authClient := authentication.NewAuthClient(config)

	ex := exporter.NewExporter(config)

	cl := cluster.NewCluster()

	wd := watchdog.NewWatchDog(config, cl)

	hcClient := healthcheck.NewHealthCheck(config, authClient, ex, wd, cl)

	// initialize api
	router := api.NewRouter(hcClient)

	// enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET"},
	})

	log.Info(fmt.Sprintf("%v", http.ListenAndServe(":2112",
		corsHandler.Handler(router)).Error()))
}
