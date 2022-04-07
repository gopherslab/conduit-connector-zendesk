package source

import (
	"conduit-connector-zendesk/config"
	"context"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
	zendesk "github.com/nukosuke/go-zendesk/zendesk"
)

type Source struct {
	sdk.UnimplementedSource
	config   config.Config
	iterator Iterator
	client   *zendesk.Client
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
	configTwo, err := config.Parse(cfg)
	if err != nil {
		return err
	}
	s.config = configTwo

	return nil
}

func (s *Source) Open(ctx context.Context, rp sdk.Position) error {
	client, err := zendesk.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to connect zendesk client:%w", err)
	}
	s.client = client
	s.client.SetSubdomain(config.ConfigKeyDomain)
	if config.ConfigKeyPassword == "" {
		s.client.SetCredential(zendesk.NewAPITokenCredential(config.ConfigKeyUserName, config.ConfigKeyToken))
	} else if config.ConfigKeyToken == "" {
		s.client.SetCredential(zendesk.NewBasicAuthCredential(config.ConfigKeyUserName, config.ConfigKeyPassword))
	} else {
		return fmt.Errorf("Enter valid credentials:%w", err)
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
	}
	return nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	sdk.Logger(ctx).Debug().Str("position", string(position)).Msg("received ack")
	return nil
}

//TODO encrypt token and password
