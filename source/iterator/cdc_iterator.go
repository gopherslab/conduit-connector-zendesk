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
	"github.com/conduitio/conduit-connector-zendesk/source/position"
)

type CDCIterator struct {
	client         *http.Client
	config         config.Config
	ticketPosition position.TicketPosition
	endOfStream    bool
	startTime      time.Time
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
	return !c.endOfStream || time.Now().Unix() > c.ticketPosition.NextIterator
}

func (c *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	var res response
	var url string

	if !c.startTime.IsZero() {
		url = fmt.Sprintf("https://%s.zendesk.com/api/v2/incremental/tickets/cursor.json?start_time=%d", c.config.Domain, c.startTime.Unix())
	}

	if c.ticketPosition.AfterURL != "" {
		url = c.ticketPosition.AfterURL
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

	c.endOfStream = res.EndOfStream

	if len(res.TicketList) == 0 {
		c.ticketPosition.NextIterator = time.Now().Add(2 * time.Minute).Unix()
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	c.ticketPosition.AfterURL = res.AfterURL
	payload, err := json.Marshal(res.TicketList)

	if err != nil {
		return sdk.Record{}, err
	}
	ticketIndex := position.TicketPosition{
		AfterURL:     c.ticketPosition.AfterURL,
		NextIterator: time.Now().Unix(),
	}

	return sdk.Record{
		Position: ticketIndex.ToRecordPosition(),
		Payload:  sdk.RawData(payload),
	}, err

}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *CDCIterator) Stop() {
	//nothing to stop
}
