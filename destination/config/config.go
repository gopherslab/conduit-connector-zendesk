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

package config

import (
	"fmt"
	"strconv"

	"github.com/conduitio/conduit-connector-zendesk/config"
)

const (
	// KeyBufferSize is the config name for buffer size.
	KeyBufferSize = "bufferSize"

	// maxBufferSize determines maximum buffer size a config can accept.
	// When config with bigger buffer size is parsed, an error is returned.
	maxBufferSize uint64 = 100
)

type Config struct {
	config.Config
	BufferSize uint64
}

// Parse validate config and configurable bufferSize
func Parse(cfg map[string]string) (Config, error) {
	defaultConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
	}

	bufferSizeString := cfg[KeyBufferSize]
	if bufferSizeString == "" {
		bufferSizeString = fmt.Sprintf("%d", maxBufferSize)
	}

	bufferSize, err := strconv.ParseUint(bufferSizeString, 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf(
			"%q config value should be a positive integer",
			KeyBufferSize,
		)
	}

	if bufferSize > maxBufferSize {
		return Config{}, fmt.Errorf(
			"%q config value should not be bigger than %d, got %d",
			KeyBufferSize,
			maxBufferSize,
			bufferSize,
		)
	}

	destinationConfig := Config{
		Config:     defaultConfig,
		BufferSize: bufferSize,
	}
	return destinationConfig, nil
}
