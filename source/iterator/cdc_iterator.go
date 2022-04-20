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
	afterURL    string
	endOfStream bool
	startTime   time.Time
}

type response struct {
	AfterCursor string        `json:"after_cursor"`
	AfterURL    string        `json:"after_url"`
	EndOfStream bool          `json:"end_of_stream"`
	TicketList  []interface{} `json:"tickets"`
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
	var res response
	var result sdk.Record
	var url string

	if !c.startTime.IsZero() {
		url = fmt.Sprintf("https://%s.zendesk.com/api/v2/incremental/tickets/cursor.json?start_time=%d", c.config.Domain, c.startTime.Unix())
	}

	if c.afterURL != "" {
		url = c.afterURL
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not access the zendesk")
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(c.config.UserName, c.config.Password))

	resp, err := c.client.Do(req)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not get the zendesk response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	ticketList, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sdk.Record{}, err
	}

	err = json.Unmarshal(ticketList, &res)
	if err != nil {
		return sdk.Record{}, err
	}

	if res.AfterURL == "" {
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	c.endOfStream = res.EndOfStream

	if len(res.TicketList) == 0 {
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	c.afterURL = res.AfterURL
	payload, err := json.Marshal(res.TicketList)

	if err != nil {
		return sdk.Record{}, err
	}

	result.Payload = sdk.RawData(payload)
	result.Position = []byte((c.afterURL))
	return result, nil

}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *CDCIterator) Stop() {
	//nothing to stop
}
