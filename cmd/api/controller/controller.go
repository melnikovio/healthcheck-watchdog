package controller

import (
	"encoding/json"
	"fmt"
	"github.com/healthcheck-exporter/cmd/healthcheck"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type ApiController struct {
	Hc *healthcheck.HealthCheck
}

// Health Ping
func (api *ApiController) Ping(w http.ResponseWriter, _ *http.Request) error {
	_, err := fmt.Fprintf(w, "pong")
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
	}

	return err
}

// Health Ping
func (api *ApiController) Health(w http.ResponseWriter, _ *http.Request) error {
	status := api.Hc.Status()
	out, err := json.Marshal(status)
	_, err = fmt.Fprintf(w, string(out))
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
	}

	return err
}
