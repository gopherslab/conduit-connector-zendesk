package config

import (
	"fmt"
	"time"
)

const (
	ConfigKeyDomain            = "domain"
	ConfigKeyUserName          = "username"
	ConfigKeyAPIToken          = "apitoken"
	ConfigKeyIterationInterval = "iterationinterval"
	DefaulIterationInterval    = "2m"
)

type Config struct {
	Domain            string
	UserName          string
	APIToken          string
	IterationInterval time.Duration
}

func Parse(cfg map[string]string) (Config, error) {

	userDomain, ok := cfg[ConfigKeyDomain]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyDomain)
	}

	userName, ok := cfg[ConfigKeyUserName]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyUserName)
	}

	userAPIToken, ok := cfg[ConfigKeyAPIToken]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyAPIToken)
	}

	itertionTime := cfg[ConfigKeyIterationInterval]
	if itertionTime == "" {
		itertionTime = DefaulIterationInterval
	}
	interval, err := time.ParseDuration(itertionTime)
	if err != nil {
		return Config{}, fmt.Errorf("%q can't parse time interval", itertionTime)
	}

	config := Config{
		Domain:            userDomain,
		UserName:          userName,
		APIToken:          userAPIToken,
		IterationInterval: interval,
	}

	return config, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
