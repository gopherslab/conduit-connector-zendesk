package config

import (
	"fmt"
	"testing"

	"github.com/conduitio/conduit-connector-zendesk/config"
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
			name: "Login with valid configuration",
			config: map[string]string{
				KeyPollingPeriod:   "5m",
				config.KeyDomain:   "testlab",
				config.KeyUserName: "test@testlab.com",
				config.KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				PollingPeriod: 300000000000,
				Config: config.Config{
					Domain:   "testlab",
					UserName: "test@testlab.com",
					APIToken: "gkdsaj)({jgo43646435#$!ga",
				},
			},
			isError: false,
		},
		{
			name: "Login with empty polling period to check default time duration",
			config: map[string]string{
				KeyPollingPeriod:   "",
				config.KeyDomain:   "testlab",
				config.KeyUserName: "test@testlab.com",
				config.KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				PollingPeriod: 120000000000,
				Config: config.Config{
					Domain:   "testlab",
					UserName: "test@testlab.com",
					APIToken: "gkdsaj)({jgo43646435#$!ga",
				},
			},
			isError: false,
		},
		{
			name: "Login without polling period",
			config: map[string]string{
				config.KeyDomain:   "testlab",
				config.KeyUserName: "test@testlab.com",
				config.KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				PollingPeriod: 120000000000,
				Config: config.Config{
					Domain:   "testlab",
					UserName: "test@testlab.com",
					APIToken: "gkdsaj)({jgo43646435#$!ga",
				},
			},
			isError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Parse(tt.config)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
				fmt.Println(tt.want, res, err)
				assert.Equal(t, res, tt.want)
			}
		})
	}
}
