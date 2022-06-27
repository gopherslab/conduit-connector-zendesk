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
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/source/iterator/mocks"
	"github.com/conduitio/conduit-connector-zendesk/source/position"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCDCIterator(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		username      string
		apiToken      string
		pollingPeriod time.Duration
		tp            position.TicketPosition
		isError       bool
	}{
		{
			name:          "NewCDCIterator with lastModifiedTime=0",
			domain:        "testlab",
			username:      "test@testlab.com",
			apiToken:      "gkdsaj)({jgo43646435#$!ga",
			pollingPeriod: time.Millisecond,
			tp:            position.TicketPosition{LastModified: time.Time{}},
		}, {
			name:          "NewCDCIterator with lastModifiedTime=2022-01-02T15:04:05Z",
			domain:        "testlab",
			username:      "test@testlab.com",
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
			res, err := NewCDCIterator(context.Background(), tt.username, tt.apiToken, tt.domain, tt.pollingPeriod, tt.tp)
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
	defer cancel()

	dummyPosition, err := (&position.TicketPosition{LastModified: time.Now(), ID: 1234}).ToRecordPosition()
	assert.NoError(t, err)
	in := sdk.Record{Position: dummyPosition}

	mockCursor := new(mocks.ZendeskCursor)
	mockCursor.On("FetchRecords", mock.Anything).Once().Return([]sdk.Record{in}, nil)

	cdc := newTestCDCIterator(ctx, t, 500*time.Millisecond, mockCursor) // half of timeout time

	out, err := cdc.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
	cdc.Stop()
	out, err = cdc.Next(ctx)
	assert.Empty(t, out)
	assert.EqualError(t, err, "iterator stopped")
}

func TestNext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	dummyPosition, err := (&position.TicketPosition{LastModified: time.Now(), ID: 1234}).ToRecordPosition()
	assert.NoError(t, err)
	in := sdk.Record{Position: dummyPosition}

	mockCursor := new(mocks.ZendeskCursor)
	mockCursor.On("FetchRecords", mock.Anything).Once().Return([]sdk.Record{in}, nil)

	cdc := newTestCDCIterator(ctx, t, 500*time.Millisecond, mockCursor) // half of timeout

	out, err := cdc.Next(ctx)
	mockCursor.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
	cancel()
	out, err = cdc.Next(ctx)
	assert.EqualError(t, err, ctx.Err().Error())
	assert.Empty(t, out)
}

func TestHasNext(t *testing.T) {
	tests := []struct {
		name          string
		fn            func(t *testing.T, c *CDCIterator, mc *mocks.ZendeskCursor)
		response      bool
		pollingPeriod time.Duration
	}{{
		name: "Has next",
		fn: func(t *testing.T, c *CDCIterator, mc *mocks.ZendeskCursor) {
			c.mux.Lock()
			defer c.mux.Unlock()
			dummyPosition, err := (&position.TicketPosition{LastModified: time.Now(), ID: 1234}).ToRecordPosition()
			assert.NoError(t, err)
			in := sdk.Record{Position: dummyPosition}
			mc.On("FetchRecords", mock.Anything).Return([]sdk.Record{in}, nil)
		},
		pollingPeriod: time.Millisecond,
		response:      true,
	}, {
		name: "no record in buffer",
		fn: func(t *testing.T, c *CDCIterator, mc *mocks.ZendeskCursor) {
			c.mux.Lock()
			defer c.mux.Unlock()
			mc.On("FetchRecords", mock.Anything).Return([]sdk.Record{}, nil)
		},
		response:      false,
		pollingPeriod: time.Millisecond,
	}, {
		name: "record in buffer, iterator stopped",
		fn: func(t *testing.T, c *CDCIterator, mc *mocks.ZendeskCursor) {
			// directly set record in buffer, to mock this scenario
			c.buffer <- sdk.Record{}
			c.Stop()
		},
		pollingPeriod: time.Millisecond,
		response:      true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			mockCursor := new(mocks.ZendeskCursor)
			cdc := newTestCDCIterator(ctx, t, tt.pollingPeriod, mockCursor)
			tt.fn(t, cdc, mockCursor)
			select {
			case <-ctx.Done():
				t.Error("timed out, did you set polling period longer than 500ms?")
			case <-time.After(tt.pollingPeriod * 2): // give iterator grace period to fetch some records
				res := cdc.HasNext(ctx)
				assert.Equal(t, tt.response, res)
				mockCursor.AssertExpectations(t)
			}
		})
	}
}

func TestStreamIterator_Stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cdc := newTestCDCIterator(ctx, t, time.Minute)
	cdc.Stop()
	rec, err := cdc.Next(ctx)
	assert.Empty(t, rec)
	assert.EqualError(t, err, "iterator stopped")
}

func newTestCDCIterator(ctx context.Context, t *testing.T, pollingPeriod time.Duration, cursors ...ZendeskCursor) *CDCIterator {
	t.Helper()
	cdc, err := NewCDCIterator(ctx, "", "", "", pollingPeriod, position.TicketPosition{}, cursors...)
	assert.NoError(t, err)
	return cdc
}
