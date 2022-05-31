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
	"testing"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-zendesk/destination"
	"github.com/conduitio/conduit-connector-zendesk/source"
	"go.uber.org/goleak"
)

func TestAcceptance(t *testing.T) {
	sourceConfig := map[string]string{
		"zendesk.domain":   "claim-bridge",
		"zendesk.userName": "pavan@claim-bridge.com",
		"zendesk.apiToken": "Xc7w2Wu5y4OlQnmGsu7MjjpF50JNxMjVQrvkQwYn",
		"pollingPeriod":    "1m",
	}
	destinationConfig := map[string]string{
		"zendesk.domain":   "claim-bridge",
		"zendesk.userName": "pavan@claim-bridge.com",
		"zendesk.apiToken": "Xc7w2Wu5y4OlQnmGsu7MjjpF50JNxMjVQrvkQwYn",
		"bufferSize":       "10",
		"maxRetries":       "1",
	}

	inputConfig := sdk.ConfigurableAcceptanceTestDriverConfig{
		Connector: sdk.Connector{
			NewSpecification: Specification,
			NewSource:        source.NewSource,
			NewDestination:   destination.NewDestination,
		},
		SourceConfig:      sourceConfig,
		DestinationConfig: destinationConfig,
		GoleakOptions:     []goleak.Option{goleak.IgnoreCurrent()},
		Skip: []string{
			// these tests are skipped, because they need a valid apiToken and empty zendesk to run properly
			// TODO: implement dummy http client to execute these with dummy data
			"TestSource_Open_ResumeAtPosition",
			"TestDestination_WriteAsync_Success",
			"TestDestination_WriteOrWriteAsync",
			"TestDestination_Write_Success",
			"TestSource_Read_Success",
			"TestSource_Read_Timeout",
		},
	}
	sdk.AcceptanceTest(t, sdk.ConfigurableAcceptanceTestDriver{
		Config: inputConfig,
	})
}
