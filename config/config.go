package config

import (
	"fmt"
)

const (
	ConfigKeyDomain        = "domain"
	ConfigKeyUserName      = "username"
	ConfigKeyPassword      = "password"
	ConfigKeyFetchInterval = "fetchinterval"
	DefaultFetchInterval   = "2"
	DefaultPerPage         = "10"
)

type Config struct {
	Domain        string
	UserName      string
	Password      string
	FetchInterval string
	PerPage       string
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

	userPassword, ok := cfg[ConfigKeyPassword]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyPassword)
	}

	config := Config{
		Domain:   userDomain,
		UserName: userName,
		Password: userPassword,
	}

	if cfg[ConfigKeyFetchInterval] != "" {
		config.FetchInterval = cfg[ConfigKeyFetchInterval]
	} else {
		config.FetchInterval = DefaultFetchInterval
	}

	return config, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
