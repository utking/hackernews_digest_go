package fetcher

import (
	"log"
	"os"

	"github.com/tkanos/gonfig"
)

const defaultConfigFile = "./config.json"

type SmtpConfig struct {
	Host     string
	Port     uint
	From     string
	Username string
	Password string
	UseTls   bool
	UseSsl   bool
	Subject  string
}

type Configuration struct {
	ApiBaseUrl     string
	DatabaseFile   string
	PurgeAfterDays uint
	Filters        []FilterItem
	EmailTo        string
	Smtp           SmtpConfig
}

func GetConfig() Configuration {
	configFile := os.Getenv("CONFIG_FILENAME")
	if configFile == "" {
		configFile = defaultConfigFile
	}

	config := Configuration{}
	err := gonfig.GetConf(configFile, &config)
	if err != nil {
		log.Fatal("CONFIG_GET: ", err)
	}
	if len(config.Filters) == 0 {
		log.Fatal("CONFIG_FILTERS no filters configured")
	}
	return config
}
