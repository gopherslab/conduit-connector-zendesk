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

package destination

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/destination/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfigure(t *testing.T) {
	invalidCfg := map[string]string{
		"zendesk.domain":   "test.lab",
		"zendesk.userName": "",
		"zendesk.apiToken": "ajgrmrop&90002p$@7",
		"pollingPeriod":    "6s",
	}

	validConfig := map[string]string{
		"zendesk.domain":   "testlab",
		"zendesk.userName": "test",
		"zendesk.apiToken": "ajgrmrop&90002p$@7",
		"pollingPeriod":    "6s",
	}

	type field struct {
		cfg map[string]string
	}
	tests := []struct {
		name    string
		field   field
		want    config.Config
		isError bool
	}{
		{
			name: "valid config",
			field: field{
				cfg: validConfig,
			},
			isError: false,
		},
		{
			name: "invalid config",
			field: field{
				cfg: invalidCfg,
			},
			isError: true,
		},
	}
	var destination Destination
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := destination.Configure(context.Background(), tt.field.cfg)
			if tt.isError {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestNewDestination(t *testing.T) {
	dest := NewDestination()
	assert.NotNil(t, dest)
}

func TestOpen(t *testing.T) {
	d := NewDestination()
	err := d.Open(context.Background())
	assert.Nil(t, err)
}

func TestWriteAsync(t *testing.T) {
	tests := []struct {
		name   string
		record sdk.Record
		ack    sdk.AckFunc
		err    error
		dest   Destination
	}{
		{
			name: "write empty record",
			record: sdk.Record{
				Payload: sdk.RawData([]byte(``)),
			},
			err: fmt.Errorf("no records from server to write"),
		},
		{
			name: "valid case",
			record: sdk.Record{
				Payload: sdk.RawData([]byte(`"dummy_data":"12345"`)),
			},
			dest: Destination{
				mux: &sync.Mutex{},
				cfg: Config{
					BufferSize: 2,
				},
				buffer:       make([]sdk.Record, 0),
				ackFuncCache: make([]sdk.AckFunc, 0),
			},
		},
		{
			name: "write invalid case with flush error",
			record: sdk.Record{
				Payload: sdk.RawData([]byte(`"dummy_data":"12345"`)),
			},
			dest: Destination{
				mux: &sync.Mutex{},
				cfg: Config{
					BufferSize: 1,
				},
				writer: func() Writer {
					w := &mocks.Writer{}
					w.On("Write", mock.Anything, mock.Anything).Return(errors.New("testing error"))
					return w
				}(),
				buffer:       make([]sdk.Record, 1),
				ackFuncCache: make([]sdk.AckFunc, 0),
			},
			err: errors.New("testing error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dest.WriteAsync(context.Background(), tt.record, tt.ack)
			if tt.err != nil {
				assert.NotNil(t, err)
				assert.Equal(t, err, tt.err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestTearDown(t *testing.T) {
	tests := []struct {
		name string
		dest Destination
		want error
	}{
		{
			name: "writer invalid case for teardown",
			dest: Destination{
				mux: &sync.Mutex{},
				cfg: Config{
					BufferSize: 2,
				},
				writer: func() Writer {
					w := &mocks.Writer{}
					w.On("Write", mock.Anything, mock.Anything).Return(errors.New("testing error"))
					return w
				}(),
				buffer:       make([]sdk.Record, 0),
				ackFuncCache: make([]sdk.AckFunc, 0),
			},
			want: errors.New("testing error"),
		},
		{
			name: "writer valid case for teardown",
			dest: Destination{
				mux: &sync.Mutex{},
				cfg: Config{
					BufferSize: 2,
				},
				writer: func() Writer {
					w := &mocks.Writer{}
					w.On("Write", mock.Anything, mock.Anything).Return(nil)
					return w
				}(),
				buffer:       make([]sdk.Record, 0),
				ackFuncCache: make([]sdk.AckFunc, 0),
				err:          nil,
			},
		},
		{
			name: "nil writer case",
			dest: Destination{
				mux: &sync.Mutex{},
				cfg: Config{
					BufferSize: 2,
				},
				buffer:       make([]sdk.Record, 0),
				ackFuncCache: make([]sdk.AckFunc, 0),
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dest.Teardown(context.Background())
			if tt.want != nil {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestFlush(t *testing.T) {
	tests := []struct {
		name string
		dest Destination
		want error
	}{
		{
			name: "invalid case",
			dest: Destination{
				mux: &sync.Mutex{},
				cfg: Config{
					BufferSize: 2,
				},
				writer: func() Writer {
					w := &mocks.Writer{}
					w.On("Write", mock.Anything, mock.Anything).Return(errors.New("testing error"))
					return w
				}(),
				buffer:       make([]sdk.Record, 0),
				ackFuncCache: make([]sdk.AckFunc, 0),
			},
			want: errors.New("testing error"),
		},
		{
			name: "valid case",
			dest: Destination{
				mux: &sync.Mutex{},
				cfg: Config{
					BufferSize: 2,
				},
				writer: func() Writer {
					w := &mocks.Writer{}
					w.On("Write", mock.Anything, mock.Anything).Return(nil)
					return w
				}(),
				buffer:       make([]sdk.Record, 0),
				ackFuncCache: make([]sdk.AckFunc, 0),
				err:          nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dest.Flush(context.Background())
			if tt.want != nil {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}
