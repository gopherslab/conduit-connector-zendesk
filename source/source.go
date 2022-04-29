package source

import (
	"context"

	"github.com/conduitio/conduit-connector-zendesk/config"
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

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	zendeskConfig, err := config.Parse(cfg)
	if err != nil {
		return err
	}
	s.config = zendeskConfig
	return nil
}

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

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	return nil
}
