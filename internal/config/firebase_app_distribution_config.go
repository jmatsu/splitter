package config

import (
	"github.com/jmatsu/splitter/internal/logger"
	"strings"
)

// FirebaseAppDistributionConfig contains the enough values to use Firebase App Distribution.
// ref: https://firebase.google.com/docs/app-distribution
type FirebaseAppDistributionConfig struct {
	serviceNameHolder `yaml:",inline"`
	ExecutionConfig   `yaml:",inline"`

	// An app ID. You can get this value from the firebase console's project setting.
	AppId string `yaml:"app-id" required:"true"`

	// Access token that has permission to use App Distribution
	AccessToken string `yaml:"access-token,omitempty"`

	// A path to credentials file. If the both of this and access token are given, access token takes priority.
	GoogleCredentialsPath string `yaml:"credentials-path" env:"GOOGLE_APPLICATION_CREDENTIALS"`

	// A list of group aliases.
	GroupAliases []string `yaml:"group-aliases,omitempty"`
}

func (c *FirebaseAppDistributionConfig) Validate() error {
	if err := validateMissingValues(c); err != nil {
		return err
	}

	if c.AccessToken == "" && c.GoogleCredentialsPath == "" {
		logger.Logger.Warn().Msg("we recommend specifying a token or credentials path explicitly")
	} else if c.AccessToken != "" && c.GoogleCredentialsPath != "" {
		logger.Logger.Warn().Msg("the specified access token is prioritized")
	}

	return nil
}

func (c *FirebaseAppDistributionConfig) ProjectNumber() string {
	// <num>:<project number>:<os>:<uid>
	return strings.SplitN(c.AppId, ":", 3)[1]
}
