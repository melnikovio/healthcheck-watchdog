package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/healthcheck-watchdog/cmd/common"
	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
)

const configPath = "config.json"

func NewConfiguration() (config *model.Config) {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Error(fmt.Sprintf("couldn't load configuration: %s", err.Error()))
		panic(err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Error(fmt.Sprintf("couldn't parse configuration: %s", err.Error()))
		panic(err)
	}

	if err = validate(config); err != nil {
		log.Error(fmt.Sprintf("Failed to validate configuration: %s", err.Error()))
		panic(err)
	}

	if err = update(config); err != nil {
		log.Error(fmt.Sprintf("Failed to update configuration: %s", err.Error()))
		panic(err)
	}

	setLogLevel(config)

	return config
}

func validate(config *model.Config) error {
	var err error
	if config == nil {
		err = errors.New("empty configuration")
		return err
	}

	if config.Jobs == nil || len(config.Jobs) == 0 {
		err = errors.New("missing monitoring tasks")
		return err
	}

	return nil
}

func update(config *model.Config) error {
	if config.AuthenticationClients == nil {
		config.AuthenticationClients = make(map[string]model.Authentication)
	}

	if !(config.Authentication == model.Authentication{}) {
		log.Warn(
			"Using deprecated configuration of Authentication. Please update Authentication section")

		_, ok := config.AuthenticationClients[common.DefaultClientId]
		if !ok {
			config.AuthenticationClients[common.DefaultClientId] = model.Authentication{
				AuthUrl:      config.Authentication.AuthUrl,
				ClientId:     config.Authentication.ClientId,
				ClientSecret: config.Authentication.ClientSecret,
			}
			log.Warn("Default client set from Authentication section")
		}
	}

	for i := range config.Jobs {
		if config.Jobs[i].AuthEnabled {
			log.Warn(
				fmt.Sprintf(
					"Using deprecated configuration of auth in job %s. Please update auth_enabled in job configuration",
					config.Jobs[i].Id))
			config.Jobs[i].Auth = model.Auth{
				Enabled: config.Jobs[i].AuthEnabled,
				Client:  common.DefaultClientId,
			}
		}

		if config.Jobs[i].MetricName == "" {
			log.Warn(
				fmt.Sprintf(
					"Please set metric name in job %s. Using id of job instead", config.Jobs[i].Id))
			config.Jobs[i].MetricName = config.Jobs[i].Id
		}
	}

	return nil
}

func setLogLevel(config *model.Config) {
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		log.Info("Seting log level from environment variable")
		config.LogLevel = v
	}

	switch strings.ToUpper(config.LogLevel) {
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "TRACE":
		log.SetLevel(log.TraceLevel)
	}
}
