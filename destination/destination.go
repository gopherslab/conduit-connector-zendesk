/*
Copyright © 2022 Meroxa, Inc.

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
	"net/http"
	"sync"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/destination/config"
	"github.com/conduitio/conduit-connector-zendesk/destination/writer"
)

type Writer interface {
	Write(ctx context.Context, records []sdk.Record) error
	Stop(ctx context.Context)
}

type Destination struct {
	sdk.UnimplementedDestination
	cfg          config.Config
	buffer       []sdk.Record
	ackFuncCache []sdk.AckFunc
	err          error
	mux          *sync.Mutex
	writer       Writer
}

func NewDestination() sdk.Destination {
	return &Destination{}
}

// Configure parses and initializes the config.
func (d *Destination) Configure(ctx context.Context, cfg map[string]string) error {
	configuration, err := config.Parse(cfg)
	if err != nil {
		return err
	}

	d.cfg = configuration
	return nil
}

// Open http client
func (d *Destination) Open(ctx context.Context) error {
	d.mux = &sync.Mutex{}
	d.buffer = make([]sdk.Record, 0, d.cfg.BufferSize)
	d.ackFuncCache = make([]sdk.AckFunc, 0, d.cfg.BufferSize)
	d.writer, d.err = writer.NewWriter(d.cfg, &http.Client{})

	return nil
}

func (d *Destination) WriteAsync(ctx context.Context, r sdk.Record, ackFunc sdk.AckFunc) error {
	if d.err != nil {
		return d.err
	}

	if len(r.Payload.Bytes()) == 0 {
		return d.err
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	d.buffer = append(d.buffer, r)
	d.ackFuncCache = append(d.ackFuncCache, ackFunc)

	if len(d.buffer) >= int(d.cfg.BufferSize) {
		err := d.Flush(ctx)
		if err != nil {
			return err
		}
	}
	return d.err
}

func (d *Destination) Flush(ctx context.Context) error {
	bufferedRecords := d.buffer
	d.buffer = d.buffer[:0]

	err := d.writer.Write(ctx, bufferedRecords)
	if err != nil {
		return err
	}

	for _, ack := range d.ackFuncCache {
		err := ack(d.err)
		if err != nil {
			return err
		}
	}
	d.ackFuncCache = d.ackFuncCache[:0]
	return nil
}

func (d *Destination) Teardown(ctx context.Context) error {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.Flush(ctx)
	d.writer.Stop(ctx)
	d.writer = nil
	return nil
}
