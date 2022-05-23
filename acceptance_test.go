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
	"testing"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/destination"
	"github.com/conduitio/conduit-connector-zendesk/source"
	"go.uber.org/goleak"
)

var records []sdk.Record
var pos sdk.Position
var ctx context.Context

func init() {
	var inputRecords []sdk.Record
	inputBytes := []byte(`{"allow_attachments":true,"allow_channelback":false,"assignee_id":393061744458,"brand_id":5030783098269,"collaborator_ids":[],"created_at":"2022-04-30T13:15:17Z","custom_fields":[],"description":"Hi there,\n\nI’m sending an email because I’m having a problem setting up your new product. Can you help me troubleshoot?\n\nThanks,\n The Customer\n\n","due_at":null,"email_cc_ids":[],"external_id":null,"fields":[],"follower_ids":[],"followup_ids":[],"forum_topic_id":null,"generated_timestamp":1651324517,"group_id":5030759730717,"has_incidents":false,"id":1,"is_public":true,"organization_id":null,"priority":"normal","problem_id":null,"raw_subject":"Sample ticket: Meet the ticket","recipient":null,"requester_id":5030783190813,"satisfaction_rating":null,"sharing_agreement_ids":[],"status":"open","subject":"Sample ticket: Meet the ticket","submitter_id":393061744458,"tags":["sample","support","zendesk"],"ticket_form_id":5030774969245,"type":"incident","updated_at":"2022-04-30T13:15:17Z","url":"https://claim-bridge.zendesk.com/api/v2/tickets/1.json","via":{"channel":"sample_ticket","source":{"from":{},"rel":null,"to":{}}}}`)
	inputRecord := sdk.Record{
		Payload:  sdk.RawData(inputBytes),
		Position: []byte(`{"LastModified":"2022-05-08T02:48:21Z","ID":1}`),
	}
	inputRecords = append(inputRecords, inputRecord)

	records = inputRecords
	pos = []byte(`{"LastModified":"2022-05-08T02:48:21Z","ID":1}`)
	ctx = context.Background()
}

func TestAcceptance(t *testing.T) {
	sourceConfig := map[string]string{
		"zendesk.domain":   "claim-bridge",
		"zendesk.userName": "pavan@claim-bridge.com",
		"zendesk.apiToken": "Xc7w2Wu5y4OlQnmGsu7MjjpF50JNxMjVQrvkQwYn",
		"pollingPeriod":    "1m",
	}
	destinationConfig := map[string]string{
		"zendesk.domain":   "claim-bridge",
		"zendesk.userName": "pavan@claim-bridge.com",
		"zendesk.apiToken": "Xc7w2Wu5y4OlQnmGsu7MjjpF50JNxMjVQrvkQwYn",
		"bufferSize":       "10",
	}

	inputConfig := sdk.ConfigurableAcceptanceTestDriverConfig{
		Connector: sdk.Connector{
			NewSpecification: Specification,
			NewSource:        source.NewSource,
			NewDestination:   destination.NewDestination,
		},
		SourceConfig:      sourceConfig,
		DestinationConfig: destinationConfig,
		GoleakOptions:     []goleak.Option{goleak.IgnoreCurrent()},
		Skip: []string{
			// these tests are skipped, because they need a valid apiToken and empty zendesk to run properly
			// TODO: implement dummy http client to execute these with dummy data
			"TestSource_Open_ResumeAtPosition",
			"TestDestination_WriteAsync_Success",
			"TestDestination_WriteOrWriteAsync",
			"TestDestination_Write_Success",
			"TestSource_Read_Success",
			"TestSource_Read_Timeout",
		},
	}
	sdk.AcceptanceTest(t, sdk.ConfigurableAcceptanceTestDriver{
		Config: inputConfig,
	})
}