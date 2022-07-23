package fetcher

import (
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

func GetConfig(filename string) (Configuration, error) {
	config := Configuration{}
	err := gonfig.GetConf(filename, &config)
	if err != nil {
		return Configuration{}, err
	}
	if len(config.Filters) == 0 {
		return Configuration{}, err
	}
	return config, nil
}
