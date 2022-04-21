package position

import (
	"encoding/json"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type TicketPosition struct {
	AfterURL     string
	NextIterator int64
}

func (pos TicketPosition) ToRecordPosition() sdk.Position {
	res, err := json.Marshal(pos)
	if err != nil {
		return sdk.Position{}
	}

	return res
}

func ParsePosition(p sdk.Position) (TicketPosition, error) {
	var err error

	if p == nil {
		return TicketPosition{}, err
	}

	var tp TicketPosition
	err = json.Unmarshal(p, &tp)
	if err != nil {
		return TicketPosition{}, fmt.Errorf("Couldn't parse the position timestamp")
	}

	return tp, err
}
