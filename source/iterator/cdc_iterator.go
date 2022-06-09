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

package iterator

import (
	"context"
	"errors"
	"sync"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/source/position"
	"github.com/conduitio/conduit-connector-zendesk/zendesk"
	"gopkg.in/tomb.v2"
)

type ZendeskCursor interface {
	FetchRecords(ctx context.Context) ([]sdk.Record, error)
}

//go:generate mockery --name=ZendeskCursor

type CDCIterator struct {
	lastModifiedTime time.Time         // ticket last updated time
	tomb             *tomb.Tomb        // new tomb
	ticker           *time.Ticker      // records time interval for next iteration
	caches           chan []sdk.Record // cache to store array of tickets
	buffer           chan sdk.Record   // buffer to store individual ticket object
	cursor           ZendeskCursor
	mux              *sync.Mutex // mux to avoid race condition while setting custom cursor
}

// NewCDCIterator will initialize CDCIterator parameters and also initialize goroutine to fetch records from server
func NewCDCIterator(
	ctx context.Context,
	username, apiToken, domain string, // config params
	pollingPeriod time.Duration,
	tp position.TicketPosition,
	cursors ...ZendeskCursor,
) (*CDCIterator, error) {
	tmbWithCtx, _ := tomb.WithContext(ctx)
	lastModified := tp.LastModified
	if lastModified.IsZero() {
		lastModified = time.Unix(0, 0)
	}

	var cursor ZendeskCursor = zendesk.NewCursor(username, apiToken, domain, lastModified)
	if len(cursors) > 0 {
		cursor = cursors[0]
	}

	cdc := &CDCIterator{
		tomb:             tmbWithCtx,
		caches:           make(chan []sdk.Record, 1),
		buffer:           make(chan sdk.Record, 1),
		ticker:           time.NewTicker(pollingPeriod),
		lastModifiedTime: lastModified,
		cursor:           cursor,
		mux:              &sync.Mutex{},
	}

	cdc.tomb.Go(cdc.startCDC(ctx))
	cdc.tomb.Go(cdc.flush)

	return cdc, nil
}

// HasNext return true when buffer is not empty
func (c *CDCIterator) HasNext(_ context.Context) bool {
	return len(c.buffer) > 0 || !c.tomb.Alive() // return true in case of go routines dying, error will be returned by Next
}

// Next will check the case whether to push data into buffer
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

// startCDC fetches records and set next position with lastmodified time of the ticket
func (c *CDCIterator) startCDC(ctx context.Context) func() error {
	return func() error {
		defer close(c.caches)
		for {
			select {
			case <-c.tomb.Dying():
				return c.tomb.Err()
			case <-c.ticker.C:
				c.mux.Lock()
				records, err := c.cursor.FetchRecords(ctx)
				c.mux.Unlock() // avoid defer, to stop locking the cursor for long duration, while it is not being used
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
				select {
				case <-c.tomb.Dying():
					return c.tomb.Err()
				case c.buffer <- record:
				}
			}
		}
	}
}

func (c *CDCIterator) Stop() {
	c.ticker.Stop()
	c.tomb.Kill(errors.New("iterator stopped"))
}
