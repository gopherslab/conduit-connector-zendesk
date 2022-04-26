package position

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type TicketPosition struct {
	AfterURL     string
	NextIterator time.Time
}

//ToRecordPosition will extract the after_url from the ticket result json
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
	//parse the next position to sdk.Record
	err = json.Unmarshal(p, &tp)
	if err != nil {
		return TicketPosition{}, fmt.Errorf("Couldn't parse the after_cursor position")
	}

	return tp, err
}
