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
package zendesk

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCursor_FetchRecords(t *testing.T) {
	th := &testHandler{
		t:          t,
		url:        &url.URL{Host: "", Path: "/api/v2/incremental/tickets/cursor.json", RawQuery: "start_time=1"},
		statusCode: 200,
		resp:       []byte(`{"after_url":"something","tickets":[{"id":1,"updated_at":"2022-05-08T05:49:55Z","created_at":"2022-05-08T05:49:55Z"}]}`),
		username:   "dummy_user",
		apiToken:   "dummy_token",
	}
	testServer := httptest.NewServer(th)
	cursor := &Cursor{
		userName:         th.username,
		apiToken:         th.apiToken,
		client:           &http.Client{},
		baseURL:          testServer.URL,
		lastModifiedTime: time.Unix(0, 0),
	}
	ctx := context.Background()
	recs, err := cursor.FetchRecords(ctx)
	assert.NoError(t, err)
	assert.Len(t, recs, 1)
}

func TestCursor_FetchRecords_RateLimit(t *testing.T) {
	// in case of nextRun being set later than now, no processing should occur
	cursor := &Cursor{
		nextRun: time.Now().Add(time.Minute),
	}
	recs, err := cursor.FetchRecords(context.Background())
	assert.Nil(t, err)
	assert.Nil(t, recs)
}

func TestCursor_FetchRecords_429(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "93")
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
	cursor := &Cursor{
		userName:         th.username,
		apiToken:         th.apiToken,
		client:           &http.Client{},
		baseURL:          testServer.URL,
		lastModifiedTime: time.Unix(0, 0),
		afterURL:         fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?cursor=some_dummy", testServer.URL),
	}
	ctx := context.Background()
	recs, err := cursor.FetchRecords(ctx)
	assert.NoError(t, err)
	assert.Len(t, recs, 0)
	assert.GreaterOrEqual(t, cursor.nextRun.Unix(), time.Now().Add(90*time.Second).Unix())
}

func TestCursor_FetchRecords_500(t *testing.T) {
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/api/v2/incremental/tickets/cursor.json", RawQuery: "cursor=some_dummy"},
		statusCode: 500,
		resp:       []byte(``),
		username:   "dummy_user",
		apiToken:   "dummy_token",
	}
	testServer := httptest.NewServer(th)
	cursor := &Cursor{
		userName:         th.username,
		apiToken:         th.apiToken,
		client:           &http.Client{},
		baseURL:          testServer.URL,
		lastModifiedTime: time.Unix(0, 0),
		afterURL:         fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?cursor=some_dummy", testServer.URL),
	}
	ctx := context.Background()
	recs, err := cursor.FetchRecords(ctx)
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
