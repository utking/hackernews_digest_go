package fetcher

import (
	"github.com/tkanos/gonfig"
)

type SmtpConfig struct {
	Host     string
	From     string
	Username string
	Password string
	Subject  string
	Port     uint
	UseTls   bool
	UseSsl   bool
}

type Database struct {
	Driver   string
	Database string
	Username string
	Password string
	Address  string
}

type Configuration struct {
	ApiBaseUrl         string
	EmailTo            string
	Filters            []FilterItem
	BlacklistedDomains []string
	Database           Database
	Smtp               SmtpConfig
	PurgeAfterDays     uint
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
