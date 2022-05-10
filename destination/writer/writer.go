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
	"github.com/conduitio/conduit-connector-zendesk/destination/destinationConfig"
	"github.com/conduitio/conduit-connector-zendesk/destination/model"
)

func Write(ctx context.Context, client *http.Client, cfg destinationConfig.Config, input []sdk.Record) error {

	var nextRun time.Time
	if nextRun.After(time.Now()) {
		return nil
	}

	bufferedTicket, err := jsonParseRecord(input)
	if err != nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, destinationConfig.ZendeskBulkImportURL, bytes.NewBuffer(bufferedTicket))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(cfg.UserName, cfg.APIToken))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not connect to zendesk client")
	}

	defer resp.Body.Close()

	// Validation for httpStatusCode 429 - Too many Requests, Retry value after `93s`
	if resp.StatusCode == http.StatusTooManyRequests {
		// NOTE: https://developer.zendesk.com/documentation/ticketing/using-the-zendesk-api/best-practices-for-avoiding-rate-limiting/#catching-errors-caused-by-rate-limiting
		retryValue, err := strconv.ParseInt(resp.Header.Get("Retry_After"), 10, 64)
		if err != nil {
			return fmt.Errorf("unable to get retry value: %w", err)
		}

		// skip hitting API till retry_after duration passes
		nextRun = time.Now().Add(time.Duration(retryValue) * time.Second)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non 200 status code received(%v)", resp.StatusCode)
	}

	return nil
}

func jsonParseRecord(buffer []sdk.Record) ([]byte, error) {
	output := model.ZdTickets{
		Tickets: make([]model.Ticket, 0),
	}

	for _, zdTicket := range buffer {
		var importTickets model.Ticket
		err := json.Unmarshal(zdTicket.Payload.Bytes(), &importTickets)
		if err != nil {
			continue
		}
		output.Tickets = append(output.Tickets, importTickets)
	}

	m, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func basicAuth(username, token string) string {
	auth := username + "/token:" + token
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
