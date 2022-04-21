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

	var interval time.Duration
	var err error

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

	if cfg[ConfigKeyIterationInterval] != "" {
		interval, err = time.ParseDuration(cfg[ConfigKeyIterationInterval])
		if err != nil {
			interval, err = time.ParseDuration(DefaulIterationInterval)
		}
	} else {
		interval, err = time.ParseDuration(DefaulIterationInterval)
	}
	if err != nil {
		return Config{}, err
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
