/*
Copyright © 2022 Meroxa, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package source

import (
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]string
		want   Config
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
				PollingPeriod: time.Minute * 5,
				Config: config.Config{
					Domain:   "testlab",
					UserName: "test@testlab.com",
					APIToken: "gkdsaj)({jgo43646435#$!ga",
				},
			},
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
				PollingPeriod: time.Second * 6,
				Config: config.Config{
					Domain:   "testlab",
					UserName: "test@testlab.com",
					APIToken: "gkdsaj)({jgo43646435#$!ga",
				},
			},
		},
		{
			name: "Login without polling period",
			config: map[string]string{
				config.KeyDomain:   "testlab",
				config.KeyUserName: "test@testlab.com",
				config.KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				PollingPeriod: time.Second * 6,
				Config: config.Config{
					Domain:   "testlab",
					UserName: "test@testlab.com",
					APIToken: "gkdsaj)({jgo43646435#$!ga",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, _ := Parse(tt.config)
			assert.NotNil(t, res)
			assert.Equal(t, res, tt.want)
		})
	}
}
