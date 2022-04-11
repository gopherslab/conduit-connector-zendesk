package iterator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	sdk "github.com/conduitio/conduit-connector-sdk"
	zendesk "github.com/nukosuke/go-zendesk/zendesk"
)

type CDCIterator struct {
	client  *zendesk.Client
	tickets []zendesk.Ticket
	page    zendesk.Page
}

//CursorOptions - Ticket audit
//initial time for cursor options- time.Now()

func NewCDCIterator(ctx context.Context, client *zendesk.Client, page int, perPage int) (*CDCIterator, error) {

	var ticketListOptions *zendesk.TicketListOptions

	ticketListOptions.PageOptions.Page = page
	ticketListOptions.PageOptions.PerPage = perPage

	ticket, resultPages, err := getAllTickets(ctx, ticketListOptions, client)

	if err != nil {
		return nil, err
	}

	cdc := &CDCIterator{
		client:  client,
		tickets: ticket,
		page:    resultPages,
	}
	return cdc, nil
}

func getAllTickets(ctx context.Context, ticketList *zendesk.TicketListOptions, client *zendesk.Client) ([]zendesk.Ticket, zendesk.Page, error) {

	ticketData, page, err := client.GetTickets(ctx, ticketList)
	if err != nil {
		log.Fatalln(err)
	}
	return ticketData, page, err
}

func (c *CDCIterator) HasNext() bool {
	return c.page.HasNext()
}

func (c *CDCIterator) Next(ctx context.Context) (*sdk.Record, error) {

	data, err := json.Marshal(c.tickets)
	if err != nil {
		return &sdk.Record{}, fmt.Errorf("could not read the object body: %w", err)
	}
	result := &sdk.Record{
		Metadata: map[string]string{
			"Page":    *c.page.NextPage,
			"PerPage": strconv.Itoa(int(c.page.Count)),
		},
		Payload: sdk.RawData(data),
	}

	return result, nil
}

func (c *CDCIterator) Stop() {

}
