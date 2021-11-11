package configuration

import (
	"encoding/json"
	"github.com/healthcheck-exporter/cmd/model"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

func GetConfiguration() (config *model.Config) {
	configFile, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Error("Couldn't load configuration")
		panic(err)
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil || config == nil {
		log.Error("Couldn't parse configuration")
		panic(err)
	}

	return config
}
