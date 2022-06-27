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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/destination"
	"github.com/conduitio/conduit-connector-zendesk/source"
	"github.com/conduitio/conduit-connector-zendesk/zendesk"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

type ticket struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

var (
	domain   string
	userName string
	apiToken string
	baseURL  string
)

func TestAcceptance(t *testing.T) {
	domain = strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_DOMAIN"))
	if domain == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_DOMAIN")
		t.FailNow()
	}

	userName = strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_USER_NAME"))
	if userName == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_USER_NAME")
		t.FailNow()
	}

	apiToken = strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_API_TOKEN"))
	if apiToken == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_API_TOKEN")
		t.FailNow()
	}

	baseURL = fmt.Sprintf("https://%s.zendesk.com", domain)

	sourceConfig := map[string]string{
		config.KeyDomain:        domain,
		config.KeyUserName:      userName,
		config.KeyAPIToken:      apiToken,
		source.KeyPollingPeriod: "1s",
	}
	destConfig := map[string]string{
		config.KeyDomain:          domain,
		config.KeyUserName:        userName,
		config.KeyAPIToken:        apiToken,
		destination.KeyBufferSize: "10",
	}

	clearTickets := func(t *testing.T) {
		assert.NoError(t, deleteTickets(t)) // archive the zendesk tickets - max 100 tickets per page
	}

	sdk.AcceptanceTest(t, AcceptanceTestDriver{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())), // nolint: gosec // only used for testing
		ConfigurableAcceptanceTestDriver: sdk.ConfigurableAcceptanceTestDriver{
			Config: sdk.ConfigurableAcceptanceTestDriverConfig{
				Connector: sdk.Connector{
					NewSpecification: Specification,
					NewSource:        source.NewSource,
					NewDestination:   destination.NewDestination,
				},
				SourceConfig:      sourceConfig,
				DestinationConfig: destConfig,
				BeforeTest: func(t *testing.T) {
				},
				GoleakOptions: []goleak.Option{
					goleak.IgnoreCurrent(),
					goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
					// keep-alive http connection, will be closed automatically in some time
					goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
				},

				// As we are fetching the deleted tickets as well, The Zendesk API may return the deleted tickets for multiple days,
				// even if we mark them for permanent deletion. If we use time.Sleep to ensure the newly created tickets are returned by the Zendesk API,
				// we can't be sure if we will only fetch the newly created tickets or if the deleted tickets will also be returned.
				// (For ex: while creating a PoC for below-mentioned steps, I received a ticket with ID 135, which had already been deleted, was not visible in console.
				// For reference, ID of latest created tickets were 735+, and we are regularly deleting the tickets after tests, in subsequent test run)
				// Regarding skipping the snapshot, We will need to add a new config startTime in source connector which can be used to start
				// fetching the tickets changes after the current time. The process will look something like below:
				//
				// - Add startTime config
				// - Use the startTime config in iterator
				// - In Acceptance tests, pass current time as startTime
				// - Write a new ticket to zendesk (to calculate the ID that will be generated for newly generated tickets)
				// - Wait for 2 minutes to ensure the newly created ticket is propagated before cleaning up the zendesk tickets and starting the acceptance tests.
				// - Overwrite the 4 skipped tests, and add 2 minutes sleep time between write and read of the tickets. (Currently we have 5 seconds context timeout)
				// - Before checking for equality of received and expected tickets, parse record position and ticket data to overwrite the time related fields, to ensure they are ignored while comparing tickets.
				// - After all the tests that write tickets to zendesk, delete all the tickets (wait 2 minutes before calling delete, wait can be skipped for the overwritten tests)
				//
				// Maybe we can write these acceptance tests in next phase when we add startTime in the config parameters.
				// https://support.zendesk.com/hc/en-us/articles/4599509725466-Removal-of-permanently-deleted-Ticket-IDs
				Skip: []string{
					"TestDestination_WriteAsync_Success",
					"TestSource_Open_ResumeAtPositionCDC",
					"TestSource_Open_ResumeAtPositionSnapshot",
					"TestSource_Read_Success",
				},
				AfterTest: func(t *testing.T) {
					clearTickets(t) // clear all tickets from zendesk
				},
			},
		},
	})
}

type AcceptanceTestDriver struct {
	rand *rand.Rand
	sdk.ConfigurableAcceptanceTestDriver
}

func (d AcceptanceTestDriver) GenerateRecord(*testing.T) sdk.Record {
	payload := fmt.Sprintf(`{"description":"%s","subject":"%s","raw_subject":"%s"}`, d.randString(32), d.randString(32), d.randString(32))
	return sdk.Record{
		Position:  sdk.Position(fmt.Sprintf(`{last_modified_time:%v,id:"%v",}`, time.Now().Add(1*time.Second), 0)),
		Metadata:  nil,
		CreatedAt: time.Time{},
		Key:       sdk.RawData(fmt.Sprintf("%v", 0)),
		Payload:   sdk.RawData(payload),
	}
}

func (d AcceptanceTestDriver) randString(n int) string {
	const letterBytes = `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	sb := strings.Builder{}
	sb.Grow(n)
	// src.Int63() generates 63 random bits, enough for letterIdxMax characters
	for i, cache, remain := n-1, d.rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = d.rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}

func deleteTickets(t *testing.T) error {
	var res ticket
	cursor := zendesk.NewCursor(userName, apiToken, domain, time.Unix(0, 0))
	ticketIDs := make([]string, 0)

	// fetching lists of ticket id to delete
	records, _ := cursor.FetchRecords(context.Background())
	for _, record := range records {
		err := json.Unmarshal(record.Payload.Bytes(), &res)
		if err != nil {
			return err
		}

		// skip deleting already deleted tickets
		if res.Status != "deleted" {
			id := fmt.Sprint(res.ID)
			ticketIDs = append(ticketIDs, id)
		}
	}

	if len(ticketIDs) != 0 {
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodDelete,
			fmt.Sprintf("%s/api/v2/tickets/destroy_many?ids=%s", baseURL, strings.Join(ticketIDs, ",")),
			nil,
		)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("Authorization", "Basic "+basicAuth(userName, apiToken))
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		client.CloseIdleConnections()
	}
	return nil
}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
