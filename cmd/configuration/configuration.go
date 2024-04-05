package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

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
		log.Error(fmt.Sprintf("Failed to initialize exporter: %s", err.Error()))
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
