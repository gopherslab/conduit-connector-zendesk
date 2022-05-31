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
package zendesk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/stretchr/testify/assert"
)

func TestNewWriter(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		userName   string
		apiToken   string
		maxRetries uint64
		isError    bool
	}{
		{
			name:       "New writer connection with configured buffer size",
			domain:     "testlab",
			userName:   "Test",
			apiToken:   "jgkfdgrjIuU78490",
			maxRetries: uint64(2),
			isError:    false,
		},
		{
			name:     "New writer with buffer with max buffersize",
			domain:   "testlab",
			userName: "Test",
			apiToken: "jgkfdgrjIuU78490",
			isError:  false,
		},
		{
			name:     "new writer without buffer configuration",
			domain:   "testlab",
			userName: "Test",
			apiToken: "jgkfdgrjIuU78490",
			isError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewBulkImporter(tt.userName, tt.apiToken, tt.domain, tt.maxRetries)
			assert.NotNil(t, res)
			assert.Equal(t, tt.userName, res.userName)
			assert.Equal(t, tt.apiToken, res.apiToken)
		})
	}
}

func TestWrite_RawDataPayload(t *testing.T) {
	ticketPayload := `{"description":"Some dummy description","priority":"normal","subject":"Sample ticket:Meet the ticket","tags":["sample","support","zendesk"],"type":"incident"}`
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/api/v2/imports/tickets/create_many"},
		statusCode: 200,
		wantBody:   fmt.Sprintf(`{"tickets":[%s]}`, ticketPayload),
		resp:       []byte(`{"job_status": {"id": "3179087242cac73b72a59df1f1dcf3df","url": "https://testlab.zendesk.com/api/v2/job_statuses/3179087242cac73b72a59df1f1dcf3df.json","total": null,"progress": null,"status": "queued","message": null,"results": null}`),
		username:   "dummy_user",
		apiToken:   "dummy_token",
	}
	testServer := httptest.NewServer(th)
	writer := &BulkImporter{
		baseURL:  testServer.URL,
		client:   &http.Client{},
		userName: th.username,
		apiToken: th.apiToken,
	}

	var inputRecords []sdk.Record
	inputRecord := sdk.Record{
		Payload: sdk.RawData(ticketPayload),
	}
	inputRecords = append(inputRecords, inputRecord)

	ctx := context.Background()
	err := writer.Write(ctx, inputRecords)
	assert.NoError(t, err)
}

func TestWrite_StructuredDataPayload(t *testing.T) {
	ticketPayload := `{"description":"Some dummy description","priority":"normal","subject":"Sample ticket: Meet the ticket","tags":["sample","support","zendesk"],"type":"incident"}`
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/api/v2/imports/tickets/create_many"},
		statusCode: 200,
		wantBody:   fmt.Sprintf(`{"tickets":[%s]}`, ticketPayload),
		resp:       []byte(`{"job_status": {"id": "3179087242cac73b72a59df1f1dcf3df","url": "https://testlab.zendesk.com/api/v2/job_statuses/3179087242cac73b72a59df1f1dcf3df.json","total": null,"progress": null,"status": "queued","message": null,"results": null}`),
		username:   "dummy_user",
		apiToken:   "dummy_token",
	}
	testServer := httptest.NewServer(th)
	writer := &BulkImporter{
		baseURL:  testServer.URL,
		client:   &http.Client{},
		userName: th.username,
		apiToken: th.apiToken,
	}

	var inputRecords []sdk.Record
	inputRecord := sdk.Record{
		Payload: sdk.StructuredData{
			"description": "Some dummy description",
			"priority":    "normal",
			"subject":     "Sample ticket: Meet the ticket",
			"tags":        []string{"sample", "support", "zendesk"},
			"type":        "incident",
		},
	}
	inputRecords = append(inputRecords, inputRecord)

	ctx := context.Background()
	err := writer.Write(ctx, inputRecords)
	assert.NoError(t, err)
}

func TestWrite_429(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "1")
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/api/v2/imports/tickets/create_many"},
		statusCode: 429,
		resp:       []byte(``),
		username:   "dummy_user",
		apiToken:   "dummy_token",
		header:     header,
	}
	testServer := httptest.NewServer(th)

	writer := &BulkImporter{
		baseURL:    testServer.URL,
		client:     &http.Client{},
		userName:   th.username,
		apiToken:   th.apiToken,
		retryCount: 1,
	}
	var inputRecords []sdk.Record
	inputBytes := []byte(`{
	"description": "Some dummy description",
	"priority": "normal",
	"subject": "Sample ticket: Meet the ticket",
	"tags": ["sample", "support", "zendesk"],
	"type": "incident"
}`)
	inputRecord := sdk.Record{
		Payload: sdk.RawData(inputBytes),
	}
	inputRecords = append(inputRecords, inputRecord)
	ctx := context.Background()
	err := writer.Write(ctx, inputRecords)
	assert.EqualError(t, err, fmt.Sprintf("rate-limit exceeded, total retries: %d", writer.retryCount))
}

func TestWrite_500(t *testing.T) {
	header := http.Header{}
	header.Set("Retry_After", "93")
	th := &testHandler{
		t:          t,
		url:        &url.URL{Host: "", Path: "/api/v2/imports/tickets/create_many"},
		statusCode: 500,
		resp:       []byte(`some_dummy_error`),
		username:   "dummy_user",
		apiToken:   "dummy_token",
		header:     header,
	}
	testServer := httptest.NewServer(th)

	writer := &BulkImporter{
		baseURL:  testServer.URL,
		client:   newHTTPClient(),
		userName: th.username,
		apiToken: th.apiToken,
	}
	var inputRecords []sdk.Record
	inputBytes := []byte(`{
	"description": "Some dummy description",
	"priority": "normal",
	"subject": "Sample ticket: Meet the ticket",
	"tags": ["sample", "support", "zendesk"],
	"type": "incident"
}`)
	inputRecord := sdk.Record{
		Payload: sdk.RawData(inputBytes),
	}
	inputRecords = append(inputRecords, inputRecord)

	ctx := context.Background()
	err := writer.Write(ctx, inputRecords)
	assert.EqualError(t, err, "non 200 status code(500) received(some_dummy_error)")
}
