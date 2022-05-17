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
package writer

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
	destinationConfig "github.com/conduitio/conduit-connector-zendesk/destination/config"

	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	t          *testing.T
	url        *url.URL
	statusCode int
	header     http.Header
	resp       []byte
	username   string
	apiToken   string
}

func TestNewWriter(t *testing.T) {
	tests := []struct {
		name    string
		config  destinationConfig.Config
		isError bool
	}{
		{
			name: "New writer connection with configured buffer size",
			config: destinationConfig.Config{
				Config: config.Config{
					Domain:   "testlab",
					UserName: "Test",
					APIToken: "jgkfdgrjIuU78490",
				},
				BufferSize: 10,
			},
			isError: false,
		},
		{
			name: "New writer with buffer with max buffersize",
			config: destinationConfig.Config{
				Config: config.Config{
					Domain:   "testlab",
					UserName: "Test",
					APIToken: "jgkfdgrjIuU78490",
				},
				BufferSize: 100,
			},
			isError: false,
		},
		{
			name: "new writer without buffer configuration",
			config: destinationConfig.Config{
				Config: config.Config{
					Domain:   "testlab",
					UserName: "Test",
					APIToken: "jgkfdgrjIuU78490",
				},
			},
			isError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewWriter(tt.config, &http.Client{})
			assert.NotNil(t, res)
			assert.Equal(t, tt.config.UserName, res.userName)
			assert.Equal(t, tt.config.APIToken, res.apiToken)
			assert.NotNil(t, tt.config.BufferSize)

		})
	}
}

func TestWrite(t *testing.T) {
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/api/v2/imports/tickets/create_many"},
		statusCode: 200,
		resp:       []byte(`{"job_status": {"id": "3179087242cac73b72a59df1f1dcf3df","url": "https://testlab.zendesk.com/api/v2/job_statuses/3179087242cac73b72a59df1f1dcf3df.json","total": null,"progress": null,"status": "queued","message": null,"results": null}`),
		username:   "dummy_user",
		apiToken:   "dummy_token",
	}
	testServer := httptest.NewServer(th)
	writer := &Writer{
		url:      testServer.URL,
		client:   &http.Client{},
		nextRun:  time.Time{},
		userName: th.username,
		apiToken: th.apiToken,
	}

	var inputRecords []sdk.Record
	inputBytes := []byte(`{"allow_attachments":true,"allow_channelback":false,"assignee_id":393061744458,"brand_id":5030783098269,"collaborator_ids":[],"created_at":"2022-04-30T13:15:17Z","custom_fields":[],"description":"Hi there,\n\nI’m sending an email because I’m having a problem setting up your new product. Can you help me troubleshoot?\n\nThanks,\n The Customer\n\n","due_at":null,"email_cc_ids":[],"external_id":null,"fields":[],"follower_ids":[],"followup_ids":[],"forum_topic_id":null,"generated_timestamp":1651324517,"group_id":5030759730717,"has_incidents":false,"id":1,"is_public":true,"organization_id":null,"priority":"normal","problem_id":null,"raw_subject":"Sample ticket: Meet the ticket","recipient":null,"requester_id":5030783190813,"satisfaction_rating":null,"sharing_agreement_ids":[],"status":"open","subject":"Sample ticket: Meet the ticket","submitter_id":393061744458,"tags":["sample","support","zendesk"],"ticket_form_id":5030774969245,"type":"incident","updated_at":"2022-04-30T13:15:17Z","url":"https://claim-bridge.zendesk.com/api/v2/tickets/1.json","via":{"channel":"sample_ticket","source":{"from":{},"rel":null,"to":{}}}}`)
	inputRecord := sdk.Record{
		Payload: sdk.RawData(inputBytes),
	}
	inputRecords = append(inputRecords, inputRecord)

	ctx := context.Background()
	err := writer.Write(ctx, inputRecords)
	assert.NoError(t, err)
}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assert.Equal(t.t, "Basic "+base64.StdEncoding.EncodeToString([]byte(t.username+"/token:"+t.apiToken)), r.Header.Get("Authorization"))

	for key, val := range t.header {
		w.Header().Set(key, val[0])
	}
	w.WriteHeader(t.statusCode)
	_, _ = w.Write(t.resp)
}

func TestWrite_429(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "93")
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

	writer := &Writer{
		url:      testServer.URL,
		client:   &http.Client{},
		nextRun:  time.Unix(0, 0),
		userName: th.username,
		apiToken: th.apiToken,
	}
	var inputRecords []sdk.Record
	inputBytes := []byte(`{"allow_attachments":true,"allow_channelback":false,"assignee_id":393061744458,"brand_id":5030783098269,"collaborator_ids":[],"created_at":"2022-04-30T13:15:17Z","custom_fields":[],"description":"Hi there,\n\nI’m sending an email because I’m having a problem setting up your new product. Can you help me troubleshoot?\n\nThanks,\n The Customer\n\n","due_at":null,"email_cc_ids":[],"external_id":null,"fields":[],"follower_ids":[],"followup_ids":[],"forum_topic_id":null,"generated_timestamp":1651324517,"group_id":5030759730717,"has_incidents":false,"id":1,"is_public":true,"organization_id":null,"priority":"normal","problem_id":null,"raw_subject":"Sample ticket: Meet the ticket","recipient":null,"requester_id":5030783190813,"satisfaction_rating":null,"sharing_agreement_ids":[],"status":"open","subject":"Sample ticket: Meet the ticket","submitter_id":393061744458,"tags":["sample","support","zendesk"],"ticket_form_id":5030774969245,"type":"incident","updated_at":"2022-04-30T13:15:17Z","url":"https://claim-bridge.zendesk.com/api/v2/tickets/1.json","via":{"channel":"sample_ticket","source":{"from":{},"rel":null,"to":{}}}}`)
	inputRecord := sdk.Record{
		Payload: sdk.RawData(inputBytes),
	}
	inputRecords = append(inputRecords, inputRecord)
	ctx := context.Background()
	err := writer.Write(ctx, inputRecords)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, writer.nextRun.Unix(), time.Now().Add(90*time.Second).Unix())
}

func TestWrite_500(t *testing.T) {
	header := http.Header{}
	header.Set("Retry_After", "93")
	th := &testHandler{
		t:          t,
		url:        &url.URL{Host: "", Path: "/api/v2/imports/tickets/create_many"},
		statusCode: 500,
		resp:       []byte(``),
		username:   "dummy_user",
		apiToken:   "dummy_token",
		header:     header,
	}
	testServer := httptest.NewServer(th)

	writer := &Writer{
		url:      testServer.URL,
		client:   &http.Client{},
		nextRun:  time.Time{},
		userName: th.username,
		apiToken: th.apiToken,
	}
	var inputRecords []sdk.Record
	inputBytes := []byte(`{"allow_attachments":true,"allow_channelback":false,"assignee_id":393061744458,"brand_id":5030783098269,"collaborator_ids":[],"created_at":"2022-04-30T13:15:17Z","custom_fields":[],"description":"Hi there,\n\nI’m sending an email because I’m having a problem setting up your new product. Can you help me troubleshoot?\n\nThanks,\n The Customer\n\n","due_at":null,"email_cc_ids":[],"external_id":null,"fields":[],"follower_ids":[],"followup_ids":[],"forum_topic_id":null,"generated_timestamp":1651324517,"group_id":5030759730717,"has_incidents":false,"id":1,"is_public":true,"organization_id":null,"priority":"normal","problem_id":null,"raw_subject":"Sample ticket: Meet the ticket","recipient":null,"requester_id":5030783190813,"satisfaction_rating":null,"sharing_agreement_ids":[],"status":"open","subject":"Sample ticket: Meet the ticket","submitter_id":393061744458,"tags":["sample","support","zendesk"],"ticket_form_id":5030774969245,"type":"incident","updated_at":"2022-04-30T13:15:17Z","url":"https://claim-bridge.zendesk.com/api/v2/tickets/1.json","via":{"channel":"sample_ticket","source":{"from":{},"rel":null,"to":{}}}}`)
	inputRecord := sdk.Record{
		Payload: sdk.RawData(inputBytes),
	}
	inputRecords = append(inputRecords, inputRecord)

	ctx := context.Background()
	err := writer.Write(ctx, inputRecords)
	assert.EqualError(t, err, "non 200 status code received(500)")
}
