package config

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {

	tests := []struct {
		name   string
		config map[string]string
		want   Config
	}{
		{
			name: "Login with basic authentication",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "12345678",
			},
			want: Config{
				Domain:   "testlab",
				UserName: "test@testlab.com",
				Password: "12345678",
			},
		},
		{
			name: "Login with api token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyToken:    "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				Domain:   "testlab",
				UserName: "test@testlab.com",
				Token:    "gkdsaj)({jgo43646435#$!ga",
			},
		},
		{
			name: "Login with oauth token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigOAuthToken:  "Okprwermnxcrt[$#09ji454",
			},
			want: Config{
				Domain:     "testlab",
				UserName:   "test@testlab.com",
				OAuthToken: "Okprwermnxcrt[$#09ji454",
			},
		},
		{
			name: "Login with basic authentication without domain",
			config: map[string]string{
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "Okprwermnxcrt[$#09ji454",
			},
			want: Config{},
		},
		{
			name: "Login with basic authentication without domain and username",
			config: map[string]string{
				ConfigKeyPassword: "Okprwermnxcrt[$#09ji454",
			},
			want: Config{},
		},
		{
			name:   "Login with basic authentication without domain, username and password",
			config: map[string]string{},
			want:   Config{},
		},
		{
			name: "Login with basic authentication without password",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "",
			},
			want: Config{},
		},
		{
			name: "Login with api token authentication without domain",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyToken:    "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{},
		},
		{
			name: "Login with api token authentication without domain and username",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyToken:    "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{},
		},
		{
			name: "Login with api token authentication without token",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyToken:    "",
			},
			want: Config{},
		},
		{
			name: "Login with api token authentication without domain, username and apitoken",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyToken:    "",
			},
			want: Config{},
		},
		{
			name: "Login with oauth token authentication without domain",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "test@testlab.com",
				ConfigOAuthToken:  "Okprwermnxcrt[$#09ji454",
			},
			want: Config{},
		},
		{
			name: "Login with oauth token authentication without username",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "",
				ConfigOAuthToken:  "Okprwermnxcrt[$#09ji454",
			},
			want: Config{},
		},
		{
			name: "Login with oauth token authentication without domain and username",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigOAuthToken:  "Okprwermnxcrt[$#09ji454",
			},
			want: Config{},
		},
		{
			name: "Login with oauth token authentication without oauth token",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigOAuthToken:  "",
			},
			want: Config{},
		},
		{
			name: "Login with oauth token authentication without domain, username and oauth token",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigOAuthToken:  "",
			},
			want: Config{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := Parse(tt.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse = %v, want = %v", got, tt.want)

			}
			got, _ := Parse(tt.config)
			fmt.Println(tt.name, got)
		})
	}
}
