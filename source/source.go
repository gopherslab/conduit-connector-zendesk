package source

import (
	"context"
	"fmt"
	"net/http"

	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/source/iterator"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Source struct {
	sdk.UnimplementedSource
	config   config.Config
	iterator Iterator
	client   *http.Client
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
	s.client = &http.Client{}

	return nil
}

func (s *Source) Open(ctx context.Context, rp sdk.Position) error {

	var err error

	s.iterator, err = iterator.NewCDCIterator(ctx, s.client, s.config)
	if err != nil {
		return fmt.Errorf("err")
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
	sdk.Logger(ctx).Info().Msg("Shutting down Zendesk Client")
	if s.iterator != nil {
		s.iterator.Stop()
		s.iterator = nil
		s.client = nil
	}
	return nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	sdk.Logger(ctx).Debug().Str("position", string(position)).Msg("received ack")
	return nil
}
