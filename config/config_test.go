/*
Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.

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
		},
		{
			name: "Login with without domain",
			config: map[string]string{
				KeyUserName: "test@testlab.com",
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
		},
		{
			name: "Login with without username",
			config: map[string]string{
				KeyDomain:   "testlab",
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
		},
		{
			name: "Login without domain and username",
			config: map[string]string{
				KeyAPIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			want:    Config{},
			isError: true,
		},
		{
			name: "Login without APIToken",
			config: map[string]string{
				KeyDomain:   "testlab",
				KeyUserName: "test@testlab.com",
			},
			want:    Config{},
			isError: true,
		},
		{
			name:    "Login without domain, username and APIToken",
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
