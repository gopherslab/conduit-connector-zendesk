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
	"net/http"
	"strconv"
	"time"

	"github.com/conduitio/conduit-connector-zendesk/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Cursor struct {
	client           *http.Client // new http client
	userName         string       // zendesk username
	apiToken         string       // zendesk apiToken
	afterURL         string       // index url for next fetch of tickets
	nextRun          time.Time    // configurable polling period to hit zendesk api
	lastModifiedTime time.Time    // ticket last updated time
	baseURL          string       // zendesk api url
}

type response struct {
	AfterURL    *string                  `json:"after_url"`     // index for to fetch next list of tickets
	EndOfStream bool                     `json:"end_of_stream"` // boolean to indicate end of ticket fetch
	TicketList  []map[string]interface{} `json:"tickets"`       // stores list of tickets
}

func NewCursor(userName, apiToken, domain string, startTime time.Time) *Cursor {
	return &Cursor{
		client:           newHTTPClient(),
		userName:         userName,
		apiToken:         apiToken,
		baseURL:          fmt.Sprintf("https://%s.zendesk.com", domain),
		lastModifiedTime: startTime,
	}
}

// FetchRecords will export tickets from zendesk api, initial start_time is set to 0
func (c *Cursor) FetchRecords(ctx context.Context) ([]sdk.Record, error) {
	if c.nextRun.After(time.Now()) {
		return nil, nil
	}

	url := fmt.Sprintf("%s/api/v2/incremental/tickets/cursor.json?start_time=%d", c.baseURL, c.lastModifiedTime.Add(time.Second).Unix()) // add one extra second, to get newer updates only

	// if after URL is available, use that
	if c.afterURL != "" {
		url = c.afterURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not access the zendesk: %w", err)
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(c.userName, c.apiToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get the zendesk response: %w", err)
	}
	defer resp.Body.Close()

	// Validation for httpStatusCode 429 - Too many Requests, Retry value after `93s`
	if resp.StatusCode == http.StatusTooManyRequests {
		// NOTE: https://developer.zendesk.com/documentation/ticketing/using-the-zendesk-api/best-practices-for-avoiding-rate-limiting/#catching-errors-caused-by-rate-limiting
		retryValue, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to get retry value: %w", err)
		}

		// skip hitting API till retry_after duration passes
		c.nextRun = time.Now().Add(time.Duration(retryValue) * time.Second)
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 status code received(%v)", resp.StatusCode)
	}

	ticketList, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading the response body: %w", err)
	}

	var res response
	err = json.Unmarshal(ticketList, &res)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling the response body: %w", err)
	}

	if res.AfterURL != nil {
		c.afterURL = *res.AfterURL
	}

	return c.toRecords(res.TicketList)
}

// convert received ticket list to sdk.Record
func (c *Cursor) toRecords(tickets []map[string]interface{}) ([]sdk.Record, error) {
	records := make([]sdk.Record, 0, len(tickets))
	lastValidModifiedTime := c.lastModifiedTime
	for _, ticket := range tickets {
		payload, err := json.Marshal(ticket)
		if err != nil {
			return nil, fmt.Errorf("error marshaling the payload: %w", err)
		}

		id, ok := ticket["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid type of id encountered: %T", ticket["id"])
		}
		updatedAt, err := time.Parse(time.RFC3339, ticket["updated_at"].(string))
		if err != nil {
			return nil, fmt.Errorf("invalid time in updated_at field: %w", err)
		}
		createdAt, err := time.Parse(time.RFC3339, ticket["created_at"].(string))
		if err != nil {
			return nil, fmt.Errorf("invalid time in created_at field: %w", err)
		}

		// there were a few records from zendesk, which had both created_at and updated_at set to 1970-01-01T00:00:00Z
		// handle such case, to ensure we don't start pulling all the records after the pause
		if updatedAt.IsZero() {
			if createdAt.IsZero() || createdAt.Before(lastValidModifiedTime) {
				updatedAt = lastValidModifiedTime
			} else {
				updatedAt = createdAt
			}
		}

		if updatedAt.After(lastValidModifiedTime) {
			lastValidModifiedTime = updatedAt
		}

		toRecordPosition, err := (&position.TicketPosition{LastModified: updatedAt, ID: id}).ToRecordPosition()
		if err != nil {
			return nil, err
		}

		records = append(records, sdk.Record{
			Position:  toRecordPosition,
			Metadata:  nil,
			CreatedAt: createdAt,
			Key:       sdk.RawData(fmt.Sprintf("%v", id)),
			Payload:   sdk.RawData(payload),
		})
	}
	return records, nil
}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second, // add timeout to ensure the Request doesn't get stuck
	}
}
