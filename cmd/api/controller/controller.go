package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/healthcheck-watchdog/cmd/common"
	"github.com/healthcheck-watchdog/cmd/manager"
	log "github.com/sirupsen/logrus"
)

type ApiController struct {
	Manager *manager.Manager
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

// Startup Status
func (api *ApiController) Startup(w http.ResponseWriter, _ *http.Request) error {
	// Return result
	w.Header().Set(common.HeaderContentType, common.ContentTypeJson)
	err := json.NewEncoder(w).Encode(true)
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return err
}

// Ready Status
func (api *ApiController) Ready(w http.ResponseWriter, _ *http.Request) error {
	status, err := api.Manager.Ready()
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// Return result
	w.Header().Set(common.HeaderContentType, common.ContentTypeJson)
	err = json.NewEncoder(w).Encode(status)
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return err
}

// Live Status
func (api *ApiController) Live(w http.ResponseWriter, _ *http.Request) error {
	status, err := api.Manager.Live()
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// Return result
	w.Header().Set(common.HeaderContentType, common.ContentTypeJson)
	err = json.NewEncoder(w).Encode(status)
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return err
}

// Health Status
func (api *ApiController) Health(w http.ResponseWriter, _ *http.Request) error {
	ready, err := api.Manager.Ready()
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	live, err := api.Manager.Live()
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// Return result
	w.Header().Set(common.HeaderContentType, common.ContentTypeJson)
	err = json.NewEncoder(w).Encode(ready && live)
	if err != nil {
		log.Error(fmt.Sprintf("The HTTP request failed with error: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return err
}
