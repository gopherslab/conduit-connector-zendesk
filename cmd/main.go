package main

import (
	source "github.com/conduitio/conduit-connector-zendesk/source"

	zendesk "github.com/conduitio/conduit-connector-zendesk"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

func main() {
	sdk.Serve(zendesk.Specification, source.NewSource, nil)
}
