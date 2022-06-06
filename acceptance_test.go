/*
Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.

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
package zendesk

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/config"
	"github.com/conduitio/conduit-connector-zendesk/destination"
	"github.com/conduitio/conduit-connector-zendesk/source"
	"github.com/matryer/is"
	"go.uber.org/goleak"
)

var (
	offset int
)

func TestAcceptance(t *testing.T) {
	domain := strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_DOMAIN"))
	if domain == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_DOMAIN")
		t.FailNow()
	}

	userName := strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_USER_NAME"))
	if userName == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_USER_NAME")
		t.FailNow()
	}

	apiToken := strings.TrimSpace(os.Getenv("CONDUIT_ZENDESK_API_TOKEN"))
	if apiToken == "" {
		t.Error("credentials not set in env CONDUIT_ZENDESK_API_TOKEN")
		t.FailNow()
	}

	sourceConfig := map[string]string{
		"domain":        domain,
		"userName":      userName,
		"apiToken":      apiToken,
		"pollingPeriod": "6s",
	}
	destConfig := map[string]string{
		"domain":     domain,
		"userName":   userName,
		"apiToken":   apiToken,
		"bufferSize": "10",
	}

	ctx := context.Background()
	conf, err := config.Parse(sourceConfig)
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	zuserName := conf.UserName
	ztoken := conf.APIToken
	baseURL := "https://https://d3v-meroxasupport.zendesk.com/"
	client := &http.Client{}
	svc, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.Header.Add("Authorization", "Basic "+basicAuth(zuserName, ztoken))
	client.Do(svc)
	sdk.AcceptanceTest(t, AcceptanceTestDriver{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		ConfigurableAcceptanceTestDriver: sdk.ConfigurableAcceptanceTestDriver{
			Config: sdk.ConfigurableAcceptanceTestDriverConfig{
				Connector: sdk.Connector{
					NewSpecification: Specification,
					NewSource:        source.NewSource,
					NewDestination:   destination.NewDestination,
				},
				SourceConfig:      sourceConfig,
				DestinationConfig: destConfig,
				BeforeTest: func(t *testing.T) {
				},
				GoleakOptions: []goleak.Option{goleak.IgnoreCurrent(), goleak.IgnoreTopFunction("internal/poll.runtime_pollWait")},
				Skip: []string{
					"TestDestination_WriteAsync_Success",
				},
				AfterTest: func(t *testing.T) {
					offset = 0
				},
			},
		},
	})
}

type AcceptanceTestDriver struct {
	rand *rand.Rand
	sdk.ConfigurableAcceptanceTestDriver
}

func (d AcceptanceTestDriver) WriteToSource(t *testing.T, recs []sdk.Record) []sdk.Record {
	if d.Connector().NewDestination == nil {
		t.Fatal("connector is missing the field NewDestination, either implement the destination or overwrite the driver method Write")
	}

	is := is.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// writing something to the destination should result in the same record
	// being produced by the source
	dest := d.Connector().NewDestination()
	// write to source and not the destination
	destConfig := d.SourceConfig(t)
	err := dest.Configure(ctx, destConfig)
	is.NoErr(err)

	err = dest.Open(ctx)
	is.NoErr(err)
	recs = d.generateRecords(len(recs))
	// try to write using WriteAsync and fallback to Write if it's not supported
	err = d.writeAsync(ctx, dest, recs)
	is.NoErr(err)

	cancel() // cancel context to simulate stop
	err = dest.Teardown(context.Background())
	is.NoErr(err)
	return recs
}

func (d AcceptanceTestDriver) generateRecords(count int) []sdk.Record {
	records := make([]sdk.Record, count)
	for i := range records {
		records[i] = d.generateRecord(offset + i)
	}
	offset += len(records)
	return records
}
func (d AcceptanceTestDriver) generateRecord(i int) sdk.Record {
	payload := fmt.Sprintf(`{"updated_at":"%s","description":"%s","subject":"%s","id":"%d"}`, time.Now(), d.randString(32), d.randString(32), 1)
	i++
	return sdk.Record{
		Position:  sdk.Position(fmt.Sprintf(`{last_modified_time:%v,id:"%v",}`, time.Now().Add(1*time.Second), i)),
		Metadata:  nil,
		CreatedAt: time.Time{},
		Key:       sdk.RawData(fmt.Sprintf("%v", i)),
		Payload:   sdk.RawData(payload),
	}
}

func (d AcceptanceTestDriver) randString(n int) string {
	const letterBytes = `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	sb := strings.Builder{}
	sb.Grow(n)
	// src.Int63() generates 63 random bits, enough for letterIdxMax characters
	for i, cache, remain := n-1, d.rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = d.rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}

// writeAsync writes records to destination using Destination.WriteAsync.
func (d AcceptanceTestDriver) writeAsync(ctx context.Context, dest sdk.Destination, records []sdk.Record) error {
	var waitForAck sync.WaitGroup
	var ackErr error

	for _, r := range records {
		waitForAck.Add(1)
		ack := func(err error) error {
			defer waitForAck.Done()
			if ackErr == nil { // only overwrite a nil error
				ackErr = err
			}
			return nil
		}
		err := dest.WriteAsync(ctx, r, ack)
		if err != nil {
			return err
		}
	}

	// flush to make sure the records get written to the destination
	err := dest.Flush(ctx)
	if err != nil {
		return err
	}

	waitForAck.Wait()
	if ackErr != nil {
		return ackErr
	}

	// records were successfully written
	return nil
}

func basicAuth(username, apiToken string) string {
	auth := username + "/token:" + apiToken
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
