package iterator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
)

type CDCIterator struct {
	client      *http.Client
	request     *http.Request
	config      config.Config
	afterCursor string
	afterURL    string
	endOfStream bool
	startTime   time.Time
}

type response struct {
	AfterCursor string `json:"after_cursor"`
	AfterURL    string `json:"after_url"`
	EndOfStream bool   `json:"end_of_stream"`
}

func NewCDCIterator(ctx context.Context, config config.Config, rp string) (*CDCIterator, error) {

	cdc := &CDCIterator{
		client:      &http.Client{},
		config:      config,
		endOfStream: false,
		startTime:   time.Unix(0, 0),
		afterURL:    rp,
	}
	return cdc, nil
}

func (c *CDCIterator) HasNext(ctx context.Context) bool {
	if !c.endOfStream {
		return true
	}
	return false
}

func (c *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	var URL string
	var res response
	var result sdk.Record

	if !c.startTime.IsZero() {
		URL = fmt.Sprintf("https://%s.zendesk.com/api/v2/incremental/tickets/cursor.json?start_time=%d", c.config.Domain, c.startTime.Unix())
	}

	if c.afterURL != "" {
		URL = c.afterURL
	}

	req, err := http.NewRequest(http.MethodGet, URL, nil)

	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not access the zendesk")
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(c.config.UserName, c.config.Password))

	resp, err := c.client.Do(req)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not get the zendesk response")
	}
	defer resp.Body.Close()

	ticketList, err := ioutil.ReadAll(resp.Body)

	json.Unmarshal(ticketList, &res)
	if res.AfterURL != "" {
		c.afterURL = res.AfterURL
		c.endOfStream = res.EndOfStream
	}

	//Writing record to conduit
	if err == nil {
		result.Payload = sdk.RawData(ticketList)
		result.Position = []byte(fmt.Sprintf("%s", c.afterURL))
	}
	return result, err
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *CDCIterator) Stop() {
	//nothing to stop
}
