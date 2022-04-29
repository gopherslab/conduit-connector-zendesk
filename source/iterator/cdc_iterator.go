package iterator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/source/position"
)

type CDCIterator struct {
	client         *http.Client
	config         config.Config
	ticketPosition position.TicketPosition
	endOfStream    bool
	startTime      time.Time
	RetryAfter     time.Duration
}

type response struct {
	AfterCursor string        `json:"after_cursor"`
	AfterURL    string        `json:"after_url"`
	EndOfStream bool          `json:"end_of_stream"`
	TicketList  []interface{} `json:"tickets"`
}

func NewCDCIterator(ctx context.Context, config config.Config, tp position.TicketPosition) (*CDCIterator, error) {

	cdc := &CDCIterator{
		client:         &http.Client{},
		config:         config,
		endOfStream:    false,
		startTime:      time.Unix(0, 0),
		ticketPosition: tp,
	}

	return cdc, nil
}

func (c *CDCIterator) HasNext(ctx context.Context) bool {
	return time.Now().After(c.ticketPosition.NextIterator)
}

func (c *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	var res response
	var url string

	// Starttime is the initial url to be hit for zendesk api
	if !c.startTime.IsZero() {
		url = fmt.Sprintf("https://%s.zendesk.com/api/v2/incremental/tickets/cursor.json?start_time=%d", c.config.Domain, c.startTime.Unix())
	}
	// Ticketposition is parsed from json response and validated
	if c.ticketPosition.AfterURL != "" {
		url = c.ticketPosition.AfterURL
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not access the zendesk")
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(c.config.UserName, c.config.APIToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not get the zendesk response")
	}
	defer resp.Body.Close()

	// Validation for httpStatusCode 429 - Too many Requests, Retry value after `93s`
	if resp.StatusCode == http.StatusTooManyRequests {
		// NOTE: https://developer.zendesk.com/documentation/ticketing/using-the-zendesk-api/best-practices-for-avoiding-rate-limiting/#catching-errors-caused-by-rate-limiting
		retryValue, err := strconv.Atoi(resp.Header.Get("Retry_After"))
		if err != nil {
			return sdk.Record{
				Position: c.ticketPosition.ToRecordPosition(),
			}, fmt.Errorf("unable to get retry value")
		}

		// IterationInterval between two successive api request to zendesk
		c.RetryAfter = time.Duration(retryValue)
		c.ticketPosition.NextIterator = time.Now().Add(c.RetryAfter)
		return sdk.Record{
			Position: c.ticketPosition.ToRecordPosition(),
		}, sdk.ErrBackoffRetry
	}

	if resp.StatusCode != http.StatusOK {
		return sdk.Record{
			Position: c.ticketPosition.ToRecordPosition(),
		}, sdk.ErrBackoffRetry
	}

	ticketList, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sdk.Record{}, err
	}

	err = json.Unmarshal(ticketList, &res)
	if err != nil {
		return sdk.Record{}, err
	}

	c.endOfStream = res.EndOfStream

	// Validate the ticket response, if zero, api request is retracted, else it will be sent
	if len(res.TicketList) == 0 {
		c.ticketPosition.NextIterator = time.Now().Add(c.config.IterationInterval)
		return sdk.Record{
			Position: c.ticketPosition.ToRecordPosition(),
		}, sdk.ErrBackoffRetry
	}

	c.ticketPosition.AfterURL = res.AfterURL
	payload, err := json.Marshal(res.TicketList)

	if err != nil {
		return sdk.Record{}, err
	}
	ticketIndex := position.TicketPosition{
		AfterURL:     c.ticketPosition.AfterURL,
		NextIterator: time.Now(),
	}

	return sdk.Record{
		Position: ticketIndex.ToRecordPosition(),
		Payload:  sdk.RawData(payload),
	}, err

}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *CDCIterator) Stop() {
	c.startTime = time.Unix(0, 0)
}
