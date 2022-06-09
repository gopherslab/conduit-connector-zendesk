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
	"io/ioutil"
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
	"go.uber.org/goleak"
)

type response struct {
	TicketList []map[string]interface{} `json:"tickets"` // stores list of tickets
}

var (
	Domain    string
	UserName  string
	ApiToken  string
	client    = http.Client{}
	ticketIDs []string
)

func TestAcceptance(t *testing.T) {
	Domain = strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_DOMAIN"))
	if Domain == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_DOMAIN")
		t.FailNow()
	}

	UserName = strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_USER_NAME"))
	if UserName == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_USER_NAME")
		t.FailNow()
	}

	ApiToken = strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_API_TOKEN"))
	if ApiToken == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_API_TOKEN")
		t.FailNow()
	}
	sourceConfig := map[string]string{
		config.KeyDomain:        Domain,
		config.KeyUserName:      UserName,
		config.KeyAPIToken:      ApiToken,
		source.KeyPollingPeriod: "1s",
	}
	destConfig := map[string]string{
		config.KeyDomain:          Domain,
		config.KeyUserName:        UserName,
		config.KeyAPIToken:        ApiToken,
		destination.KeyBufferSize: "10",
	}

	clearTickets := func(t *testing.T) {
		err := BulkDelete()
		if err != nil {
			t.Errorf("error archiving zendesk tickets: %v", err.Error())
		}

		err = PermanentDelete()
		if err != nil {
			t.Errorf("error deleting zendesk tickets: %v", err.Error())
		}
	}

	clearTickets(t)
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
				Skip: []string{
					"TestDestination_WriteAsync_Success",
					"TestSource_Open_ResumeAtPositionCDC",
					"TestSource_Open_ResumeAtPositionSnapshot",
					"TestSource_Read_Success",
				},
				AfterTest: func(t *testing.T) {
					clearTickets(t)
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
	payload := fmt.Sprintf(`{"description":"%s","subject":"%s"}`, d.randString(32), d.randString(32))
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

func BulkDelete() error {
	ticketIDs, err := getTickets()
	BaseURL := fmt.Sprintf("https://%s.zendesk.com", Domain)
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodDelete,
		fmt.Sprintf("%s/api/v2/tickets/destroy_many?ids=%s", BaseURL, strings.Join(ticketIDs, ",")), nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(UserName, ApiToken))
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("got error response when archiving records in zendesk %w", err)
	}

	defer resp.Body.Close()

	return nil
}

func PermanentDelete() error {
	BaseURL := fmt.Sprintf("https://%s.zendesk.com", Domain)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		fmt.Sprintf("%s/api/v2/deleted_tickets/destroy_many?ids=%s", BaseURL, strings.Join(ticketIDs, ",")),
		nil)

	if err != nil {
		return fmt.Errorf("unable to delete record in zendesk server %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(UserName, ApiToken))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("got error response when permanently deleting records in zendesk %w", err)
	}

	defer resp.Body.Close()

	return nil
}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func getTickets() ([]string, error) {
	BaseURL := fmt.Sprintf("https://%s.zendesk.com", Domain)
	url := fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?start_time=%d", BaseURL, time.Now().Unix())

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(UserName, ApiToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	ticketList, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res response
	err = json.Unmarshal(ticketList, &res)
	if err != nil {
		return nil, err
	}

	for _, ticket := range res.TicketList {
		id := fmt.Sprint(ticket["id"])
		ticketIDs = append(ticketIDs, id)
	}

	return ticketIDs, nil
}
