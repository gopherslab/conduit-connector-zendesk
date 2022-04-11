package config

import (
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
			name: "Login with basic authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "Okprwermnxcrt[$#09ji454",
			},
			want: Config{
				Domain:   "",
				UserName: "test@testlab.com",
				Password: "Okprwermnxcrt[$#09ji454",
			},
		},
		{
			name: "Login with basic authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyPassword: "Okprwermnxcrt[$#09ji454",
			},
			want: Config{
				Domain:   "",
				UserName: "",
				Password: "Okprwermnxcrt[$#09ji454",
			},
		},
		{
			name: "Login with basic authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyPassword: "",
			},
			want: Config{
				Domain:   "",
				UserName: "",
				Password: "",
			},
		},
		{
			name: "Login with basic authentication",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "",
				ConfigKeyPassword: "Okprwermnxcrt[$#09ji454",
			},
			want: Config{
				Domain:   "testlab",
				UserName: "",
				Password: "Okprwermnxcrt[$#09ji454",
			},
		},
		{
			name: "Login with api token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyToken:    "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				Domain:   "",
				UserName: "test@testlab.com",
				Token:    "gkdsaj)({jgo43646435#$!ga",
			},
		},
		{
			name: "Login with api token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyToken:    "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				Domain:   "",
				UserName: "",
				Token:    "gkdsaj)({jgo43646435#$!ga",
			},
		},
		{
			name: "Login with api token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "",
				ConfigKeyToken:    "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				Domain:   "testlab",
				UserName: "",
				Token:    "gkdsaj)({jgo43646435#$!ga",
			},
		},
		{
			name: "Login with api token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyToken:    "",
			},
			want: Config{
				Domain:   "",
				UserName: "",
				Token:    "",
			},
		},
		{
			name: "Login with oauth token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "test@testlab.com",
				ConfigOAuthToken:  "Okprwermnxcrt[$#09ji454",
			},
			want: Config{
				Domain:     "",
				UserName:   "test@testlab.com",
				OAuthToken: "Okprwermnxcrt[$#09ji454",
			},
		},
		{
			name: "Login with oauth token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "",
				ConfigOAuthToken:  "Okprwermnxcrt[$#09ji454",
			},
			want: Config{
				Domain:     "testlab",
				UserName:   "",
				OAuthToken: "Okprwermnxcrt[$#09ji454",
			},
		},
		{
			name: "Login with oauth token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigOAuthToken:  "Okprwermnxcrt[$#09ji454",
			},
			want: Config{
				Domain:     "",
				UserName:   "",
				OAuthToken: "Okprwermnxcrt[$#09ji454",
			},
		},
		{
			name: "Login with oauth token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigOAuthToken:  "",
			},
			want: Config{
				Domain:     "",
				UserName:   "",
				OAuthToken: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := Parse(tt.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse = %v, want = %v", got, tt.want)
			}
		})
	}
}
