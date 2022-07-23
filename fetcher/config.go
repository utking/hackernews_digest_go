package fetcher

import (
	"github.com/tkanos/gonfig"
)

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

type Database struct {
	Driver   string
	Database string
	Username string
	Password string
	Address  string
}

type Configuration struct {
	ApiBaseUrl     string
	Database       Database
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
