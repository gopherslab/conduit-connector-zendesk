package main

import (
	zendesk "conduit-connector-zendesk"
	destination "conduit-connector-zendesk/destination"
	source "conduit-connector-zendesk/source"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

func main() {
	sdk.Serve(zendesk.Specification, source.NewSource, destination.NewDestination)
}
