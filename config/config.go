package config

import (
	"errors"
	"fmt"
)

const (
	ConfigKeyDomain   = "zendesk.domain"
	ConfigKeyUserName = "zendesk.username"
	ConfigKeyPassword = "zendesk.password"
	ConfigKeyToken    = "zendesk.token"
	ConfigOAuthToken  = "zendesk.oauthtoken"
)

type Config struct {
	Domain     string
	UserName   string
	Password   string
	Token      string
	OAuthToken string
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

	config := Config{
		Domain:   userDomain,
		UserName: userName,
	}

	userPassword, ok := cfg[ConfigKeyPassword]
	if ok {
		config.Password = userPassword
		return config, nil
	}

	userToken, ok := cfg[ConfigKeyToken]
	if ok {
		config.Token = userToken
		return config, nil
	}

	userOAuthToken, ok := cfg[ConfigOAuthToken]
	if ok {
		config.OAuthToken = userOAuthToken
		return config, nil
	}

	return Config{}, errors.New("enter valid credentials for zendesk")
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
