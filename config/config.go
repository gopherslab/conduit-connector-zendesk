package config

import "fmt"

const (
	ConfigKeyDomain   = "zendesk.domain"
	ConfigKeyUserName = "zendesk.username"
	ConfigKeyPassword = "zendesk.password"
	ConfigKeyToken    = "zendesk.token"
)

type Config struct {
	Domain   string
	UserName string
	Password string
	Token    string
}

// Parse the username, password and token from the user input
//TODO: Oauth token and refresh token implementation

func Parse(cfg map[string]string) (Config, error) {
	userDomain, ok := cfg[ConfigKeyDomain]

	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyDomain)
	}
	userName, ok := cfg[ConfigKeyUserName]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyDomain)

	}
	config := Config{
		Domain:   userDomain,
		UserName: userName,
	}
	userPassword, ok := cfg[ConfigKeyPassword]
	if ok {
		config.Password = userPassword
	}
	userToken, ok := cfg[ConfigKeyToken]

	if ok {
		config.Token = userToken
	}

	return config, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
