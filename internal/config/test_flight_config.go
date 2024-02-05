package config

import (
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
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

	if c.Password == "" && (c.ApiKey == "" || c.IssuerID == "") {
		return errors.New("app-specific password or a pair of api key and issuer id is required")
	}

	if c.Password != "" {
		if c.ApiKey != "" && c.IssuerID != "" {
			logger.Logger.Warn().Msg("api key and issuer id will be chosen for TestFlight deployment")
			c.ApiKey = ""
		} else {
			c.ApiKey = ""
			c.IssuerID = ""
		}
	}

	return nil
}
