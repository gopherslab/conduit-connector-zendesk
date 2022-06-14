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
	"github.com/conduitio/conduit-connector-zendesk/source/iterator"
	"github.com/conduitio/conduit-connector-zendesk/zendesk"
	"go.uber.org/goleak"
)

type response struct {
	TicketList map[string]interface{} `json:"tickets"` // stores list of tickets
}

var (
	domain       string
	userName     string
	apiToken     string
	client       = http.Client{}
	ticketIDs    []string
	baseURL      string
	cursor       iterator.ZendeskCursor
	lastModified time.Time
)

func TestAcceptance(t *testing.T) {
	domain = strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_DOMAIN"))
	if domain == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_DOMAIN")
		t.FailNow()
	}

	baseURL = fmt.Sprintf("https://%s.zendesk.com", domain)

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
		err := bulkDelete() // archive the zendesk tickets - max 100 tickets per page
		if err != nil {
			t.Errorf("error archiving zendesk tickets: %v", err.Error())
		}
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
				GoleakOptions: []goleak.Option{goleak.IgnoreCurrent(), goleak.IgnoreTopFunction("internal/poll.runtime_pollWait")},
				// tests skipped as the zendesk deletion takes 90 days to complete the delete lifecycle.
				// As discussed, scope of incremental flow export constraint with use of exclude_delete flag
				// test account displays the scrubbed tickets
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

func bulkDelete() error {
	var res response
	cursor = zendesk.NewCursor(userName, apiToken, domain, time.Unix(0, 0))
	ticketIDs = nil
	// fetching lists of ticket id to delete
	ticketList, _ := cursor.FetchRecords(context.Background())
	for _, ticket := range ticketList {
		err := json.Unmarshal(ticket.Payload.Bytes(), &res.TicketList)
		if err != nil {
			return err
		}

		if fmt.Sprint(res.TicketList["status"]) != "deleted" {
			id := fmt.Sprint(res.TicketList["id"])
			ticketIDs = append(ticketIDs, id)
		}
	}

	if len(ticketIDs) != 0 {
		// uses bulkdelete to archive zendesk tickets, takes comma separated ids
		callDelete(ticketIDs, baseURL)
	}
	return nil
}

func callDelete(list []string, baseURL string) {
	req, _ := http.NewRequestWithContext(
		context.Background(), http.MethodDelete,
		fmt.Sprintf("%s/api/v2/tickets/destroy_many?ids=%s", baseURL, strings.Join(list, ",")), nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(userName, apiToken))
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	// permanently deletes zendesk ticket from ticket archives
	permanentDelete(list, baseURL)
}

func permanentDelete(list []string, baseURL string) {
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		fmt.Sprintf("%s/api/v2/deleted_tickets/destroy_many?ids=%s", baseURL, strings.Join(list, ",")),
		nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(userName, apiToken))
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
