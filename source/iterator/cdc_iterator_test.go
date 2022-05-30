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
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/source/position"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"
)

func TestNewCDCIterator(t *testing.T) {
	tests := []struct {
		name          string
		Domain        string
		userName      string
		apiToken      string
		pollingPeriod time.Duration
		tp            position.TicketPosition
		isError       bool
	}{
		{
			name:          "NewCDCIterator with lastModifiedTime=0",
			Domain:        "testlab",
			userName:      "test@testlab.com",
			apiToken:      "gkdsaj)({jgo43646435#$!ga",
			pollingPeriod: time.Millisecond,
			tp:            position.TicketPosition{LastModified: time.Time{}},
		}, {
			name:          "NewCDCIterator with lastModifiedTime=2022-01-02T15:04:05Z",
			Domain:        "testlab",
			userName:      "test@testlab.com",
			apiToken:      "gkdsaj)({jgo43646435#$!ga",
			pollingPeriod: time.Millisecond,
			tp: position.TicketPosition{
				LastModified: time.Date(2022, 01, 02,
					15, 04, 05, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewCDCIterator(context.Background(), tt.userName, tt.apiToken, tt.Domain, tt.pollingPeriod, tt.tp)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
				assert.NotNil(t, res.caches)
				assert.NotNil(t, res.buffer)
				assert.NotNil(t, res.tomb)
				assert.NotNil(t, res.ticker)
				expectedTime := tt.tp.LastModified.Unix()
				if expectedTime < 0 {
					expectedTime = 0
				}
				assert.Equal(t, expectedTime, res.lastModifiedTime.Unix())
			}
		})
	}
}

func TestFlush(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	tmbWithCtx, _ := tomb.WithContext(ctx)
	cdc := &CDCIterator{
		buffer: make(chan sdk.Record, 1),
		caches: make(chan []sdk.Record, 1),
		tomb:   tmbWithCtx,
	}
	randomErr := errors.New("random error")
	cdc.tomb.Go(cdc.flush)

	in := sdk.Record{Position: []byte("some_position")}
	cdc.caches <- []sdk.Record{in}
	for {
		select {
		case <-cdc.tomb.Dying():
			assert.EqualError(t, cdc.tomb.Err(), randomErr.Error())
			cancel()
			return
		case out := <-cdc.buffer:
			assert.Equal(t, in, out)
			cdc.tomb.Kill(randomErr)
		}
	}
}

func TestNext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	tmbWithCtx, _ := tomb.WithContext(ctx)
	cdc := &CDCIterator{
		buffer: make(chan sdk.Record, 1),
		caches: make(chan []sdk.Record, 1),
		tomb:   tmbWithCtx,
	}
	cdc.tomb.Go(cdc.flush)

	in := sdk.Record{Position: []byte("some_position")}
	cdc.caches <- []sdk.Record{in}
	out, err := cdc.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
	cancel()
	out, err = cdc.Next(ctx)
	assert.EqualError(t, err, ctx.Err().Error())
	assert.Empty(t, out)
}

func TestHasNext(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(c *CDCIterator)
		response bool
	}{{
		name: "Has next",
		fn: func(c *CDCIterator) {
			c.buffer <- sdk.Record{}
		},
		response: true,
	}, {
		name:     "no record in buffer",
		fn:       func(c *CDCIterator) {},
		response: false,
	}, {
		name: "record in buffer, tomb dead",
		fn: func(c *CDCIterator) {
			c.tomb.Kill(errors.New("random error"))
			c.buffer <- sdk.Record{}
		},
		response: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cdc := &CDCIterator{buffer: make(chan sdk.Record, 1), tomb: &tomb.Tomb{}}
			tt.fn(cdc)
			res := cdc.HasNext(context.Background())
			assert.Equal(t, res, tt.response)
		})
	}
}

func TestStreamIterator_Stop(t *testing.T) {
	cdc := &CDCIterator{
		tomb:   &tomb.Tomb{},
		ticker: time.NewTicker(time.Second),
	}
	cdc.Stop()
	assert.False(t, cdc.tomb.Alive())
}
