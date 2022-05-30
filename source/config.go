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

package source

import (
	"fmt"
	"time"

	"github.com/conduitio/conduit-connector-zendesk/config"
)

const (
	KeyPollingPeriod = "pollingPeriod"

	// KeyPollingPeriod determines polling time from config, if it empty or if config not provided.
	// then the defaultPollingPeriod taken as 2 minutes.
	defaultPollingPeriod = "6s"
)

type Config struct {
	config.Config
	PollingPeriod time.Duration // time interval for next zendesk api hit
}

// Parse validate zendesk config and pollingPeriod
func Parse(cfg map[string]string) (Config, error) {
	defaultConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
	}

	pollingPeriod := cfg[KeyPollingPeriod]
	if pollingPeriod == "" {
		pollingPeriod = defaultPollingPeriod
	}
	duration, err := time.ParseDuration(pollingPeriod)
	if err != nil {
		return Config{}, fmt.Errorf("%q can't parse time interval: %w", pollingPeriod, err)
	}

	sourceConfig := Config{
		Config:        defaultConfig,
		PollingPeriod: duration,
	}
	return sourceConfig, nil
}
