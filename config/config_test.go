package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]string
		want    Config
		isError bool
	}{
		{
			name: "Login with all authentication parameters",
			config: map[string]string{
				ConfigKeyDomain:        "testlab",
				ConfigKeyUserName:      "test@testlab.com",
				ConfigKeyPassword:      "gkdsaj)({jgo43646435#$!ga",
				ConfigKeyFetchInterval: "10",
			},
			want: Config{
				Domain:        "testlab",
				UserName:      "test@testlab.com",
				Password:      "gkdsaj)({jgo43646435#$!ga",
				FetchInterval: "10",
			},
			isError: false,
		},
		{
			name: "Login with all authentication parameters with default fetch interval",
			config: map[string]string{
				ConfigKeyDomain:        "testlab",
				ConfigKeyUserName:      "test@testlab.com",
				ConfigKeyPassword:      "gkdsaj)({jgo43646435#$!ga",
				ConfigKeyFetchInterval: "",
			},
			want: Config{
				Domain:        "testlab",
				UserName:      "test@testlab.com",
				Password:      "gkdsaj)({jgo43646435#$!ga",
				FetchInterval: "2",
			},
			isError: false,
		},
		{
			name: "Login with without domain",
			config: map[string]string{
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
		},
		{
			name: "Login with without username",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyPassword: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
		},
		{
			name: "Login without domain and username",
			config: map[string]string{
				ConfigKeyPassword: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
		},
		{
			name: "Login without password",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
			},
			want:    Config{},
			isError: true,
		},
		{
			name:    "Login without domain, username and password",
			config:  map[string]string{},
			want:    Config{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Parse(tt.config)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, res, tt.want)
			}
		})
	}
}
