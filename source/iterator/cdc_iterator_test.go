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
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/source/position"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"
)

func TestNewCDCIterator(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		tp      position.TicketPosition
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
			tp: position.TicketPosition{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewCDCIterator(context.Background(), tt.config, tt.tp)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, tt.config.UserName, res.userName)
				assert.Equal(t, tt.config.APIToken, res.apiToken)
				assert.Equal(t, res.baseURL, "https://testlab.zendesk.com")
				assert.NotNil(t, res.caches)
				assert.NotNil(t, res.buffer)
				assert.NotNil(t, res.tomb)
				assert.NotNil(t, res.ticker)
			}
		})
	}
}

func TestFlush(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	tmbWithCtx, _ := tomb.WithContext(ctx)
	cdc := &CDCIterator{
		buffer: make(chan sdk.Record, 1),
		caches: make(chan []sdk.Record, 1),
		tomb:   tmbWithCtx,
	}
	randomErr := errors.New("random error")
	cdc.tomb.Go(cdc.flush)

	in := sdk.Record{Position: []byte("some_position")}
	cdc.caches <- []sdk.Record{in}
	for {
		select {
		case <-cdc.tomb.Dying():
			assert.EqualError(t, cdc.tomb.Err(), randomErr.Error())
			cancel()
			return
		case out := <-cdc.buffer:
			assert.Equal(t, in, out)
			cdc.tomb.Kill(randomErr)
		}
	}
}

func TestFetchRecords(t *testing.T) {
	th := &testHandler{
		t:          t,
		url:        &url.URL{Host: "", Path: "/api/v2/incremental/tickets/cursor.json", RawQuery: "start_time=1"},
		statusCode: 200,
		resp:       []byte(`{"after_url":"something","tickets":[{"id":1,"updated_at":"2022-05-08T05:49:55Z","created_at":"2022-05-08T05:49:55Z"}]}`),
		username:   "dummy_user",
		apiToken:   "dummy_token",
	}
	testServer := httptest.NewServer(th)
	cdc := &CDCIterator{
		userName:         th.username,
		apiToken:         th.apiToken,
		client:           &http.Client{},
		baseURL:          testServer.URL,
		lastModifiedTime: time.Unix(0, 0),
	}
	ctx := context.Background()
	recs, err := cdc.fetchRecords(ctx)
	assert.NoError(t, err)
	assert.Len(t, recs, 1)
}

func TestFetchRecords_429(t *testing.T) {
	header := http.Header{}
	header.Set("Retry_After", "93")
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/api/v2/incremental/tickets/cursor.json", RawQuery: "cursor=some_dummy"},
		statusCode: 429,
		resp:       []byte(``),
		username:   "dummy_user",
		apiToken:   "dummy_token",
		header:     header,
	}
	testServer := httptest.NewServer(th)
	cdc := &CDCIterator{
		userName:         th.username,
		apiToken:         th.apiToken,
		client:           &http.Client{},
		baseURL:          testServer.URL,
		lastModifiedTime: time.Unix(0, 0),
		afterURL:         fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?cursor=some_dummy", testServer.URL),
	}
	ctx := context.Background()
	recs, err := cdc.fetchRecords(ctx)
	assert.NoError(t, err)
	assert.Len(t, recs, 0)
	assert.GreaterOrEqual(t, cdc.nextRun.Unix(), time.Now().Add(90*time.Second).Unix())
}

func TestFetchRecords_500(t *testing.T) {
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/api/v2/incremental/tickets/cursor.json", RawQuery: "cursor=some_dummy"},
		statusCode: 500,
		resp:       []byte(``),
		username:   "dummy_user",
		apiToken:   "dummy_token",
	}
	testServer := httptest.NewServer(th)
	cdc := &CDCIterator{
		userName:         th.username,
		apiToken:         th.apiToken,
		client:           &http.Client{},
		baseURL:          testServer.URL,
		lastModifiedTime: time.Unix(0, 0),
		afterURL:         fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?cursor=some_dummy", testServer.URL),
	}
	ctx := context.Background()
	recs, err := cdc.fetchRecords(ctx)
	assert.EqualError(t, err, "non 200 status code received(500)")
	assert.Len(t, recs, 0)
}

type testHandler struct {
	t          *testing.T
	url        *url.URL
	statusCode int
	header     http.Header
	resp       []byte
	username   string
	apiToken   string
}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assert.Equal(t.t, t.url.Path, r.URL.Path)
	assert.Equal(t.t, t.url.RawQuery, r.URL.RawQuery)

	assert.Equal(t.t, "Basic "+base64.StdEncoding.EncodeToString([]byte(t.username+"/token:"+t.apiToken)), r.Header.Get("Authorization"))
	for key, val := range t.header {
		w.Header().Set(key, val[0])
	}
	w.WriteHeader(t.statusCode)
	_, _ = w.Write(t.resp)
}

func TestNext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	tmbWithCtx, _ := tomb.WithContext(ctx)
	cdc := &CDCIterator{
		buffer: make(chan sdk.Record, 1),
		caches: make(chan []sdk.Record, 1),
		tomb:   tmbWithCtx,
	}
	cdc.tomb.Go(cdc.flush)

	in := sdk.Record{Position: []byte("some_position")}
	cdc.caches <- []sdk.Record{in}
	out, err := cdc.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
	cancel()
	out, err = cdc.Next(ctx)
	assert.EqualError(t, err, ctx.Err().Error())
	assert.Empty(t, out)
}

func TestHasNext(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(c *CDCIterator)
		response bool
	}{{
		name: "Has next",
		fn: func(c *CDCIterator) {
			c.buffer <- sdk.Record{}
		},
		response: true,
	}, {
		name:     "no record in buffer",
		fn:       func(c *CDCIterator) {},
		response: false,
	}, {
		name: "record in buffer, tomb dead",
		fn: func(c *CDCIterator) {
			c.tomb.Kill(errors.New("random error"))
			c.buffer <- sdk.Record{}
		},
		response: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cdc := &CDCIterator{buffer: make(chan sdk.Record, 1), tomb: &tomb.Tomb{}}
			tt.fn(cdc)
			res := cdc.HasNext(context.Background())
			assert.Equal(t, res, tt.response)
		})
	}
}
