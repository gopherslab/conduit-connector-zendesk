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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/destination/config"
	"github.com/conduitio/conduit-connector-zendesk/destination/model"
)

type Writer struct {
	url      string       // url constructed for import tickets to zendesk
	userName string       // userName to login as admin to zendesk
	apiToken string       // token to authenticate the user
	nextRun  time.Time    // cool period for the client to hit zendesk api
	client   *http.Client // http client to connect zendesk
}

// NewWriter initialize writer to import ticket
func NewWriter(cfg config.Config, client *http.Client) (*Writer, error) {
	return &Writer{
		url:      fmt.Sprintf("https://%s.zendesk.com/api/v2/imports/tickets/create_many", cfg.Domain),
		nextRun:  time.Time{},
		client:   client,
		userName: cfg.UserName,
		apiToken: cfg.APIToken,
	}, nil
}

// Write buffer data to zendesk
func (w *Writer) Write(ctx context.Context, records []sdk.Record) error {
	if w.nextRun.After(time.Now()) {
		return nil
	}

	bufferedTicket, err := jsonParseRecord(records)
	if err != nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewBuffer(bufferedTicket))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(w.userName, w.apiToken))

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not connect to zendesk client")
	}

	defer resp.Body.Close()
	// Validation for httpStatusCode 429 - Too many Requests, Retry value after `93s`
	if resp.StatusCode == http.StatusTooManyRequests {
		// NOTE: https://developer.zendesk.com/documentation/ticketing/using-the-zendesk-api/best-practices-for-avoiding-rate-limiting/#catching-errors-caused-by-rate-limiting
		retryValue, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
		if err != nil {
			return fmt.Errorf("unable to get retry value: %w", err)
		}

		// skip hitting API till retry_after duration passes
		w.nextRun = time.Now().Add(time.Duration(retryValue) * time.Second)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non 200 status code received(%v)", resp.StatusCode)
	}

	return nil
}

func (w *Writer) Stop(_ context.Context) {
	w.client = nil
}

func jsonParseRecord(records []sdk.Record) ([]byte, error) {
	output := model.CreateManyRequest{
		Tickets: make([]model.Ticket, 0, len(records)),
	}

	for _, record := range records {
		var ticket model.Ticket
		err := json.Unmarshal(record.Payload.Bytes(), &ticket)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling the payload into ticket type: %w", err)
		}
		output.Tickets = append(output.Tickets, ticket)
	}

	m, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("error marshaling the ticket list: %w", err)
	}
	return m, nil
}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
