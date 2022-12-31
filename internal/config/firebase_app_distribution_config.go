package config

import (
	"github.com/jmatsu/splitter/internal/logger"
	"strings"
)

type FirebaseAppDistributionConfig struct {
	AccessToken           string `json:"access-token,omitempty"`
	GoogleCredentialsPath string `json:"credentials-path,omitempty" env:"GOOGLE_APPLICATION_CREDENTIALS"`
	AppId                 string `json:"app-id,omitempty" required:"true"`
}

func (c *FirebaseAppDistributionConfig) Validate() error {
	if err := validateMissingValues(c); err != nil {
		return err
	}

	if c.AccessToken == "" && c.GoogleCredentialsPath == "" {
		logger.Logger.Warn().Msg("we recommend specifying a token or credentials path explicitly")
	} else if c.AccessToken != "" && c.GoogleCredentialsPath != "" {
		logger.Logger.Warn().Msg("the both of firebase token and google credentials path are specified")
	}

	return nil
}

func (c *FirebaseAppDistributionConfig) ProjectNumber() string {
	// <num>:<project number>:<os>:<uid>
	return strings.SplitN(c.AppId, ":", 3)[1]
}
