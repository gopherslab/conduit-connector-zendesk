package destinationConfig

import (
	"fmt"
	"strconv"

	"github.com/conduitio/conduit-connector-zendesk/config"
)

const (
	// ConfigKeyBufferSize is the config name for buffer size.
	ConfigKeyBufferSize = "bufferSize"

	// MaxBufferSize determines maximum buffer size a config can accept.
	// When config with bigger buffer size is parsed, an error is returned.
	MaxBufferSize uint64 = 100

	// DefaultBufferSize is the value BufferSize assumes when the config omits
	// the buffer size parameter
	DefaultBufferSize uint64 = 2
)

type Config struct {
	config.Config
	BufferSize uint64
}

func Parse(cfg map[string]string) (Config, error) {
	defaultConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
	}

	bufferSizeString, exists := cfg[ConfigKeyBufferSize]
	if !exists || bufferSizeString == "" {
		bufferSizeString = fmt.Sprintf("%d", DefaultBufferSize)
	}

	bufferSize, err := strconv.ParseUint(bufferSizeString, 10, 32)

	if err != nil {
		return Config{}, fmt.Errorf(
			"%q config value should be a positive integer",
			ConfigKeyBufferSize,
		)
	}

	if bufferSize > MaxBufferSize {
		return Config{}, fmt.Errorf(
			"%q config value should not be bigger than %d, got %d",
			ConfigKeyBufferSize,
			MaxBufferSize,
			bufferSize,
		)
	}

	destinationConfig := Config{
		Config:     defaultConfig,
		BufferSize: bufferSize,
	}
	return destinationConfig, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
