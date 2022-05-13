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

package position

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type TicketPosition struct {
	LastModified time.Time
	ID           float64 // two tickets can have the same update time, id is to keep the position unique across tickets
}

// ToRecordPosition will extract the after_url from the ticket result json
func (pos *TicketPosition) ToRecordPosition() sdk.Position {
	res, err := json.Marshal(pos)
	if err != nil {
		return sdk.Position{}
	}

	return res
}

func ParsePosition(p sdk.Position) (TicketPosition, error) {
	var err error

	if len(p) == 0 {
		return TicketPosition{}, fmt.Errorf("ticket position is empty :%w", err)
	}

	var tp TicketPosition
	// parse the next position to sdk.Record
	err = json.Unmarshal(p, &tp)
	if err != nil {
		return TicketPosition{}, fmt.Errorf("couldn't parse the after_cursor position: %w", err)
	}

	return tp, err
}
