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

package zendesk

import (
	"github.com/conduitio/conduit-connector-zendesk/config"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

func Specification() sdk.Specification {
	return sdk.Specification{
		Name:    "zendesk",
		Summary: "An zendesk source and destination plugin for Conduit, written in Go",
		Version: "v0.1.0",
		Author:  "Meroxa,Inc.",
		SourceParams: map[string]sdk.Parameter{
			config.ConfigKeyDomain: {
				Default:     "",
				Required:    true,
				Description: "A domain is referred as the organization name to which zendesk is registered",
			},
			config.ConfigKeyUserName: {
				Default:     "",
				Required:    true,
				Description: "Login to zendesk performed using username",
			},
			config.ConfigKeyAPIToken: {
				Default:     "",
				Required:    true,
				Description: "password to login",
			},
			config.ConfigKeyIterationInterval: {
				Default:     "2m",
				Required:    false,
				Description: "Fetch interval for consecutive iterations",
			},
		},
		DestinationParams: map[string]sdk.Parameter{},
	}
}
