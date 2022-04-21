package iterator

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/source/position"
	"github.com/stretchr/testify/assert"
)

func TestNewCDCIterator(t *testing.T) {

	tests := []struct {
		name    string
		config  config.Config
		tp      position.TicketPosition
		want    *CDCIterator
		isError bool
	}{
		{
			name: "NewCDCIterator with startTime=0",
			config: config.Config{
				Domain:   "testlab",
				UserName: "test@testlab.com",
				APIToken: "gkdsaj)({jgo43646435#$!ga",
			},
			tp: position.TicketPosition{
				AfterURL:     "https://testlab.zendesk.com/api/v2/incremental/tickets/cursor.json?cursor=MTY1MDM3NzAzNS4wfHwyNnw%3D",
				NextIterator: 1650458827,
			},
			want: &CDCIterator{
				client: &http.Client{},
				config: config.Config{
					Domain:   "testlab",
					UserName: "test@testlab.com",
					APIToken: "gkdsaj)({jgo43646435#$!ga",
				},
				endOfStream: false,
				startTime:   time.Unix(0, 0),
				ticketPosition: position.TicketPosition{
					AfterURL:     "https://testlab.zendesk.com/api/v2/incremental/tickets/cursor.json?cursor=MTY1MDM3NzAzNS4wfHwyNnw%3D",
					NextIterator: 1650458827,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewCDCIterator(context.Background(), tt.config, tt.tp)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, res, tt.want)
			}
		})
	}

}

func TestHasNext(t *testing.T) {
	var cdc CDCIterator
	cdc.endOfStream = false
	tests := struct {
		name     string
		response bool
	}{
		name:     "Has next",
		response: true,
	}
	t.Run(tests.name, func(t *testing.T) {
		res := cdc.HasNext(context.Background())
		assert.Equal(t, res, tests.response)
	})
}
