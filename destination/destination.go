package destination

import (
	"context"
	"net/http"
	"sync"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/destination/destinationConfig"
	"github.com/conduitio/conduit-connector-zendesk/destination/writer"
)

type Destination struct {
	sdk.UnimplementedDestination
	Config       destinationConfig.Config
	Buffer       []sdk.Record
	AckFuncCache []sdk.AckFunc
	Client       *http.Client
	Error        error
	Mutex        *sync.Mutex
}

func NewDestination() sdk.Destination {
	return &Destination{}
}

// Configure parses and initializes the config.
func (d *Destination) Configure(ctx context.Context, cfg map[string]string) error {
	configuration, err := destinationConfig.Parse(cfg)
	if err != nil {
		return err
	}

	d.Config = configuration
	return nil
}

// Open http client
func (d *Destination) Open(ctx context.Context) error {

	d.Mutex = &sync.Mutex{}
	d.Buffer = make([]sdk.Record, 0, 1)
	d.AckFuncCache = make([]sdk.AckFunc, 0, 1)
	d.Client = &http.Client{}
	return nil
}

func (d *Destination) WriteAsync(ctx context.Context, r sdk.Record, ackFunc sdk.AckFunc) error {

	if d.Error != nil {
		return d.Error
	}

	if len(r.Payload.Bytes()) == 0 {
		return nil
	}

	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	d.Buffer = append(d.Buffer, r)
	d.AckFuncCache = append(d.AckFuncCache, ackFunc)

	//fmt.Println("default buffer size", int(destinationConfig.DefaultBufferSize))

	if len(d.Buffer) >= int(destinationConfig.DefaultBufferSize) {
		err := d.Flush(ctx)
		if err != nil {
			return err
		}
	}
	return d.Error
}

func (d *Destination) Flush(ctx context.Context) error {

	bufferedRecords := d.Buffer
	d.Buffer = d.Buffer[:0]

	err := writer.Write(ctx, d.Client, d.Config, bufferedRecords)
	if err != nil {
		return err
	}

	for _, ack := range d.AckFuncCache {
		err := ack(d.Error)
		if err != nil {
			return err
		}
	}
	d.AckFuncCache = d.AckFuncCache[:0]
	return nil
}

func (d *Destination) TearDown(_ context.Context) error {
	d.Client = nil
	return nil
}
