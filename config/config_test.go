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

package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]string
		want    Config
		isError bool
		err     error
	}{
		{
			name: "Login with all authentication parameters",
			config: map[string]string{
				KeyDomain:   "testlab",
				KeyUserName: "test@testlab.com",
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				Domain:   "testlab",
				UserName: "test@testlab.com",
				APIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			isError: false,
			err:     nil,
		},
		{
			name: "Login with all authentication parameters with default fetch interval",
			config: map[string]string{
				KeyDomain:   "testlab",
				KeyUserName: "test@testlab.com",
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want: Config{
				Domain:   "testlab",
				UserName: "test@testlab.com",
				APIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			isError: false,
			err:     nil,
		},
		{
			name: "Login with without domain",
			config: map[string]string{
				KeyUserName: "test@testlab.com",
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
			err:     fmt.Errorf("\"zendesk.domain\" config value must be set"),
		},
		{
			name: "Login with without username",
			config: map[string]string{
				KeyDomain:   "testlab",
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
			err:     fmt.Errorf("\"zendesk.userName\" config value must be set"),
		},
		{
			name: "Login without domain and username",
			config: map[string]string{
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
			err:     fmt.Errorf("\"zendesk.domain\" config value must be set"),
		},
		{
			name: "Login without APIToken",
			config: map[string]string{
				KeyDomain:   "testlab",
				KeyUserName: "test@testlab.com",
			},
			want:    Config{},
			isError: true,
			err:     fmt.Errorf("\"zendesk.apiToken\" config value must be set"),
		},
		{
			name:    "Login without domain, username and APIToken",
			config:  map[string]string{},
			want:    Config{},
			isError: true,
			err:     fmt.Errorf("\"zendesk.domain\" config value must be set"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Parse(tt.config)
			if tt.isError {
				assert.NotNil(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, res, tt.want)
			}
		})
	}
}
