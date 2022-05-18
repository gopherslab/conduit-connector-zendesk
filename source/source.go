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

package source

import (
	"context"
	"fmt"

	"github.com/conduitio/conduit-connector-zendesk/source/config"
	"github.com/conduitio/conduit-connector-zendesk/source/iterator"
	"github.com/conduitio/conduit-connector-zendesk/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Source struct {
	sdk.UnimplementedSource
	config   config.Config
	iterator Iterator
}

type Iterator interface {
	HasNext(ctx context.Context) bool
	Next(ctx context.Context) (sdk.Record, error)
	Stop()
}

func NewSource() sdk.Source {
	return &Source{}
}

// Configure parses zendsesk config
func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	zendeskConfig, err := config.Parse(cfg)
	if err != nil {
		return err
	}
	s.config = zendeskConfig
	return nil
}

// Open prepare the plugin to start sending records from the given position
func (s *Source) Open(ctx context.Context, rp sdk.Position) error {
	ticketPos, err := position.ParsePosition(rp)
	if err != nil {
		return err
	}

	s.iterator, err = iterator.NewCDCIterator(ctx, s.config, ticketPos)
	if err != nil {
		return err
	}
	return nil
}

// Read gets the next object from the zendesk api
func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	if !s.iterator.HasNext(ctx) {
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	r, err := s.iterator.Next(ctx)
	if err != nil {
		return sdk.Record{}, err
	}
	return r, nil
}

func (s *Source) Teardown(ctx context.Context) error {
	sdk.Logger(ctx).Info().Msg("shutting down zendesk client")
	if s.iterator != nil {
		s.iterator.Stop()
		s.iterator = nil
	}
	return nil
}

func (s *Source) Ack(ctx context.Context, pos sdk.Position) error {
	ticketPos, err := position.ParsePosition(pos)
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}
	sdk.Logger(ctx).Info().
		Float64("id", ticketPos.ID).
		Time("update_time", ticketPos.LastModified).
		Msg("ack received")
	return nil
}
