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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type CreateManyRequest struct {
	Tickets []Ticket `json:"tickets"`
}

// Ticket is the DS representing the zendesk ticket.
type Ticket struct {
	ID                  int           `json:"id"`
	ExternalID          interface{}   `json:"external_id"`
	Via                 Via           `json:"via"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	Type                interface{}   `json:"type"`
	Subject             string        `json:"subject"`
	RawSubject          string        `json:"raw_subject"`
	Description         string        `json:"description"`
	Priority            interface{}   `json:"priority"`
	Status              string        `json:"status"`
	Recipient           interface{}   `json:"recipient"`
	RequesterID         int64         `json:"requester_id"`
	SubmitterID         int64         `json:"submitter_id"`
	AssigneeID          int64         `json:"assignee_id"`
	OrganizationID      int64         `json:"organization_id"`
	CollaboratorIds     []interface{} `json:"collaborator_ids"`
	FollowerIds         []interface{} `json:"follower_ids"`
	EmailCcIds          []interface{} `json:"email_cc_ids"`
	ForumTopicID        interface{}   `json:"forum_topic_id"`
	ProblemID           interface{}   `json:"problem_id"`
	HasIncidents        bool          `json:"has_incidents"`
	IsPublic            bool          `json:"is_public"`
	DueAt               interface{}   `json:"due_at"`
	Tags                []interface{} `json:"tags"`
	CustomFields        []interface{} `json:"custom_fields"`
	SatisfactionRating  interface{}   `json:"satisfaction_rating"`
	SharingAgreementIds []interface{} `json:"sharing_agreement_ids"`
	Fields              []interface{} `json:"fields"`
	FollowupIds         []interface{} `json:"followup_ids"`
	TicketFormID        int64         `json:"ticket_form_id"`
	AllowChannelback    bool          `json:"allow_channelback"`
	AllowAttachments    bool          `json:"allow_attachments"`
	GeneratedTimestamp  int           `json:"generated_timestamp"`
}

type Via struct {
	Channel string `json:"channel"`
	Source  Source `json:"source"`
}

type Source struct {
	From interface{} `json:"from"`
	To   interface{} `json:"to"`
	Rel  interface{} `json:"rel"`
}

type BulkImporter struct {
	userName   string       // userName to login as admin to zendesk
	apiToken   string       // token to authenticate the user
	client     *http.Client // http client to connect zendesk
	maxRetries uint64       // max API retries in case of 429, before returning error
	baseURL    string       // zendesk api url
	retryCount uint64       // number of retry count made for current data
}

// NewBulkImporter initialize bulk importer to write bulk tickets to zendesk
func NewBulkImporter(userName, apiToken, domain string, maxRetries uint64) *BulkImporter {
	return &BulkImporter{
		//url:        fmt.Sprintf("https://%s.zendesk.com/api/v2/imports/tickets/create_many", domain),
		client:     newHTTPClient(),
		userName:   userName,
		apiToken:   apiToken,
		baseURL:    fmt.Sprintf("https://%s.zendesk.com", domain),
		maxRetries: maxRetries,
	}
}

// Write buffer data to zendesk
func (b *BulkImporter) Write(ctx context.Context, records []sdk.Record) error {
	bufferedTicket, err := parseRecords(records)
	if err != nil {
		return fmt.Errorf("unable to parse the records %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v2/imports/tickets/create_many", b.baseURL),
		bytes.NewBuffer(bufferedTicket),
	)

	if err != nil {
		return fmt.Errorf("unable to send to zendesk server %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(b.userName, b.apiToken))

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to fetch response from zendesk server %w", err)
	}

	defer resp.Body.Close()
	// Validation for httpStatusCode 429 - Too many Requests, Retry value after `93s`
	if resp.StatusCode == http.StatusTooManyRequests {
		// NOTE: https://developer.zendesk.com/documentation/ticketing/using-the-zendesk-api/best-practices-for-avoiding-rate-limiting/#catching-errors-caused-by-rate-limiting
		retryValue, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
		if err != nil {
			return fmt.Errorf("unable to get retry value: %w", err)
		}

		sdk.Logger(ctx).Trace().Int64("Retry-After", retryValue).Msg("rate limit exceeded, will retry after `Retry-After` duration")

		if b.retryCount >= b.maxRetries {
			return fmt.Errorf("rate-limit exceeded, total retries: %d", b.retryCount)
		}

		b.retryCount++
		// retry writing after the cool off duration passes, block till then
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(retryValue) * time.Second):
			return b.Write(ctx, records)
		}
	}

	// reset the retry count, in case of non 429 response.
	b.retryCount = 0

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non 200 status code received(%v)", resp.StatusCode)
	}

	return nil
}

// parseRecords unmarshal the payload data from records to Ticket struct
// and returns a marshalled CreateManyRequest, to be used to write multiple tickets to zendesk
func parseRecords(records []sdk.Record) ([]byte, error) {
	output := CreateManyRequest{
		Tickets: make([]Ticket, 0, len(records)),
	}

	for _, record := range records {
		var ticket Ticket
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
