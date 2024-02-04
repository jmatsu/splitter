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

	// Api Key
	ApiKey string `yaml:"api-key,omitempty"`

	// Issuer ID of the specified api key
	IssuerID string `yaml:"issuer-id,omitempty"`
}

func (c *TestFlightConfig) Validate() error {
	if err := validateMissingValues(c); err != nil {
		return err
	}

	if c.AppleID == "" {
		logger.Logger.Warn().Msg("we recommend specifying an AppleID explicitly")
	}

	if c.Password != "" {
		if c.ApiKey != "" || c.IssuerID != "" {
			logger.Logger.Info().Msg("password will be chosen for TestFlight deployment")
		}
	}

	return nil
}
