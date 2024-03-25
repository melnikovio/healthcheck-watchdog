package configuration

import (
	"encoding/json"
	"fmt"
	"os"

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
	if err != nil || config == nil {
		log.Error(fmt.Sprintf("couldn't parse configuration: %s", err.Error()))
		panic(err)
	}

	setLogLevel()

	return config
}

func setLogLevel() {
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		switch v {
		case "ERROR":
			log.SetLevel(log.ErrorLevel)
		case "INFO":
			log.SetLevel(log.InfoLevel)
		case "TRACE":
			log.SetLevel(log.TraceLevel)
		}
	}
}
