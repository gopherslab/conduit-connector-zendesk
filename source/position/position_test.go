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
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/stretchr/testify/assert"
)

func TestToRecordPosition(t *testing.T) {
	pos := TicketPosition{
		LastModified: time.Now(),
		ID:           0,
	}
	t.Run("Record position valid case", func(t *testing.T) {
		res, _ := pos.ToRecordPosition()
		assert.NotNil(t, res)
	})
}

func TestParsePosition(t *testing.T) {
	tests := []struct {
		name    string
		pos     sdk.Position
		want    TicketPosition
		isError bool
	}{
		{
			name: "Ticket position for valid case",
			pos:  []byte(`{"LastModified":"2022-05-08T02:48:21Z","ID":87}`),
		},
		{
			want: TicketPosition{
				ID: 87,
			},
			isError: false,
		},
		{
			name: "Ticket position for not valid case",
			pos:  []byte{},
		},
		{
			want:    TicketPosition{},
			isError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ParsePosition(tt.pos)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
			}
		})
	}
}
