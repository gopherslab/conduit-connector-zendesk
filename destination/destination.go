package desination

import (
	"github.com/conduitio/conduit-connector-zendesk/config"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Destination struct {
	sdk.UnimplementedDestination
	Buffer       []sdk.Record
	AckFuncCache []sdk.AckFunc
	Config       config.Config
}

func NewDestination() sdk.Destination {
	return &Destination{}
}
