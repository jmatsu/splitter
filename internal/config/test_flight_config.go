package config

import (
	"github.com/jmatsu/splitter/internal/logger"
)

// TestFlightConfig contains the enough values to use TestFlight.
type TestFlightConfig struct {
	serviceNameHolder `yaml:",inline"`
	ExecutionConfig   `yaml:",inline"`

	// An Apple ID.
	AppleID string `yaml:"apple-id" required:"true"`

	// App-specific password
	Password string `yaml:"password,omitempty"`
}

func (c *TestFlightConfig) Validate() error {
	if err := validateMissingValues(c); err != nil {
		return err
	}

	if c.AppleID == "" {
		logger.Logger.Warn().Msg("we recommend specifying an AppleID explicitly")
	}

	return nil
}
