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
			name: "Login with api token authentication",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				Domain:   "testlab",
				UserName: "test@testlab.com",
				Password: "gkdsaj)({jgo43646435#$!ga",
			},
		},
		{
			name: "Login with api token without domain",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{},
		},
		{
			name: "Login with api token without domain and username",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyPassword: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{},
		},
		{
			name: "Login with api token without token",
			config: map[string]string{
				ConfigKeyDomain:   "testlab",
				ConfigKeyUserName: "test@testlab.com",
				ConfigKeyPassword: "",
			},
			want: Config{},
		},
		{
			name: "Login with api token without domain, username and apitoken",
			config: map[string]string{
				ConfigKeyDomain:   "",
				ConfigKeyUserName: "",
				ConfigKeyPassword: "",
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
