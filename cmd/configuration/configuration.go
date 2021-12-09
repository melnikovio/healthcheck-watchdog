package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
)

func NewConfiguration() (config *model.Config) {
	configFile, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Error(fmt.Sprintf("couldn't load configuration: %s", err.Error()))
		panic(err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil || config == nil {
		log.Error(fmt.Sprintf("Couldn't parse configuration: %s", err.Error()))
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
