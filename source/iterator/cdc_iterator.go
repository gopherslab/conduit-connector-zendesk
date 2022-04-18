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

//StartTime - Initial entry for the zendesk cursor incremental exports
var StartTime = time.Unix(0, 0)

type CDCIterator struct {
	client      *http.Client
	request     *http.Request
	config      config.Config
	afterCursor string
	afterURL    string
	endOfStream bool
}

type response struct {
	AfterCursor string `json:"after_cursor"`
	AfterURL    string `json:"after_url"`
}

func NewCDCIterator(ctx context.Context, client *http.Client, config config.Config) (*CDCIterator, error) {

	cdc := &CDCIterator{
		client: client,
		config: config,
	}
	return cdc, nil
}

func (c *CDCIterator) HasNext(ctx context.Context) bool {

	if !StartTime.IsZero() || c.afterURL != "" {
		return true
	}
	return false
}

func (c *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {

	var URL string
	var res response
	var result sdk.Record
	URL = c.afterURL

	if !StartTime.IsZero() {
		URL = fmt.Sprintf("https://%s.zendesk.com/api/v2/incremental/tickets/cursor.json?start_time=%d", c.config.Domain, StartTime.Unix())
		StartTime = time.Unix(0, 0)
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

	//Writing record to sdk.Data
	if err == nil {
		result.Payload = sdk.RawData(ticketList)
	}

	json.Unmarshal(ticketList, &res)
	if res.AfterURL != "" {
		c.afterURL = res.AfterURL
	}

	return result, err
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *CDCIterator) Stop() {
	//nothing to stop
	if !c.endOfStream {
		c.client = nil
	}
}
