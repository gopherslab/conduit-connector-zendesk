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

package iterator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/source/position"
	"gopkg.in/tomb.v2"
)

type CDCIterator struct {
	client           *http.Client
	userName         string
	apiToken         string
	afterURL         string
	nextRun          time.Time
	lastModifiedTime time.Time
	tomb             *tomb.Tomb
	ticker           *time.Ticker
	caches           chan []sdk.Record
	buffer           chan sdk.Record
	baseURL          string
}

type response struct {
	AfterURL    *string                  `json:"after_url"`
	EndOfStream bool                     `json:"end_of_stream"`
	TicketList  []map[string]interface{} `json:"tickets"`
}

func NewCDCIterator(ctx context.Context, config config.Config, tp *position.TicketPosition) (*CDCIterator, error) {
	tmbWithCtx, ctx := tomb.WithContext(ctx)
	cdc := &CDCIterator{
		userName:         config.UserName,
		apiToken:         config.APIToken,
		client:           &http.Client{},
		lastModifiedTime: tp.LastModified,
		tomb:             tmbWithCtx,
		caches:           make(chan []sdk.Record, 1),
		buffer:           make(chan sdk.Record, 1),
		ticker:           time.NewTicker(config.PollingPeriod),
		baseURL:          fmt.Sprintf("https://%s.zendesk.com/api/v2/incremental/tickets/cursor.json", config.Domain),
	}

	cdc.tomb.Go(cdc.startCDC(ctx))
	cdc.tomb.Go(cdc.flush)

	return cdc, nil
}

func (c *CDCIterator) startCDC(ctx context.Context) func() error {
	return func() error {
		defer close(c.caches)
		for {
			select {
			case <-c.tomb.Dying():
				return c.tomb.Err()
			case <-c.ticker.C:
				records, err := c.fetchRecords(ctx)
				if err != nil {
					return err
				}
				if len(records) == 0 {
					continue
				}
				select {
				case c.caches <- records:
					pos, err := position.ParsePosition(records[len(records)-1].Position)
					if err != nil {
						return err
					}
					c.lastModifiedTime = pos.LastModified
				case <-c.tomb.Dying():
					return c.tomb.Err()
				}
			}
		}
	}
}

func (c *CDCIterator) flush() error {
	defer close(c.buffer)
	for {
		select {
		case <-c.tomb.Dying():
			return c.tomb.Err()
		case cache := <-c.caches:
			for _, record := range cache {
				c.buffer <- record
			}
		}
	}
}

func (c *CDCIterator) HasNext(_ context.Context) bool {
	return len(c.buffer) > 0 || !c.tomb.Alive() // return true in case of go routines dying, error will be returned by Next
}

func (c *CDCIterator) fetchRecords(ctx context.Context) ([]sdk.Record, error) {
	if c.nextRun.After(time.Now()) {
		return nil, nil
	}

	url := fmt.Sprintf("%s?start_time=%d", c.baseURL, c.lastModifiedTime.Add(time.Second).Unix()) // add one extra second, to get newer updates only

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
		retryValue, err := strconv.Atoi(resp.Header.Get("Retry_After"))
		if err != nil {
			return nil, fmt.Errorf("unable to get retry value: %w", err)
		}

		// skip hitting API till retry_after duration passes
		c.nextRun = time.Now().Add(time.Duration(retryValue))
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

	return toRecords(res.TicketList)
}

func toRecords(tickets []map[string]interface{}) ([]sdk.Record, error) {
	records := make([]sdk.Record, 0, len(tickets))
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
		lastModifiedTime := updatedAt
		if updatedAt.IsZero() {
			lastModifiedTime = createdAt
		}

		records = append(records, sdk.Record{
			Position:  (&position.TicketPosition{LastModified: lastModifiedTime, ID: id}).ToRecordPosition(),
			Metadata:  nil,
			CreatedAt: createdAt,
			Key:       sdk.RawData(fmt.Sprintf("%v", id)),
			Payload:   sdk.RawData(payload),
		})
	}
	return records, nil
}

func (c *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	select {
	case rec := <-c.buffer:
		return rec, nil
	case <-c.tomb.Dying():
		return sdk.Record{}, c.tomb.Err()
	case <-ctx.Done():
		return sdk.Record{}, ctx.Err()
	}
}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *CDCIterator) Stop() {
	c.ticker.Stop()
	c.tomb.Kill(errors.New("iterator stopped"))
}
