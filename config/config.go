package config

import (
	"fmt"
	"strconv"
)

const (
	ConfigKeyDomain        = "domain"
	ConfigKeyUserName      = "username"
	ConfigKeyPassword      = "password"
	ConfigKeyFetchInterval = "fetchinterval"
	DefaultFetchInterval   = int64(2)
)

type Config struct {
	Domain        string
	UserName      string
	Password      string
	FetchInterval int64
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

	fetchInterval, ok := cfg[ConfigKeyFetchInterval]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyPassword)
	}

	config := Config{
		Domain:        userDomain,
		UserName:      userName,
		Password:      userPassword,
		FetchInterval: stringToInt64(fetchInterval),
	}

	return config, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}

func stringToInt64(interval string) int64 {
	nextInterval, err := strconv.ParseInt(interval, 10, 64)
	if err != nil {
		return DefaultFetchInterval
	}
	return nextInterval
}
