package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/healthcheck-watchdog/cmd/common"
	"github.com/healthcheck-watchdog/cmd/healthcheck"
	log "github.com/sirupsen/logrus"
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

// Health Status
func (api *ApiController) Health(w http.ResponseWriter, _ *http.Request) error {
	status, err := api.Hc.Status()
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// Return result
	w.Header().Set(common.HeaderContentType, common.ContentTypeJson)
	json.NewEncoder(w).Encode(status)

	return err
}

// Health Status
func (api *ApiController) Ready(w http.ResponseWriter, _ *http.Request) error {
	status, err := api.Hc.Status()
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// Return result
	w.Header().Set(common.HeaderContentType, common.ContentTypeJson)
	json.NewEncoder(w).Encode(status)

	return err
}
