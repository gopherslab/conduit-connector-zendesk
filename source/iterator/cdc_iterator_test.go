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

package iterator

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/source/position"
	"github.com/stretchr/testify/assert"
)

func TestNewCDCIterator(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		tp      *position.TicketPosition
		want    *CDCIterator
		isError bool
	}{
		{
			name: "NewCDCIterator with lastModifiedTime=0",
			config: config.Config{
				Domain:        "testlab",
				UserName:      "test@testlab.com",
				APIToken:      "gkdsaj)({jgo43646435#$!ga",
				PollingPeriod: time.Millisecond,
			},
			tp: &position.TicketPosition{},
			want: &CDCIterator{
				client:           &http.Client{},
				lastModifiedTime: time.Unix(0, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewCDCIterator(context.Background(), tt.config, tt.tp)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, res, tt.want)
			}
		})
	}
}

func TestHasNext(t *testing.T) {
	var cdc CDCIterator
	tests := struct {
		name     string
		response bool
	}{
		name:     "Has next",
		response: true,
	}
	t.Run(tests.name, func(t *testing.T) {
		res := cdc.HasNext(context.Background())
		assert.Equal(t, res, tests.response)
	})
}
