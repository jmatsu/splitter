package config

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"strings"
)

// AppIdFragmentCount is the number of the fragments of <num>:<project number>:<os>:<uid>.
const AppIdFragmentCount = 4

// AppIdFragment returns the index-th fragment of an app id, or an empty string if the app id does
// not follow the format. FirebaseAppDistributionConfig#Validate rejects such an app id beforehand.
func AppIdFragment(appId string, index int) string {
	fragments := strings.SplitN(appId, ":", AppIdFragmentCount)

	if len(fragments) < AppIdFragmentCount {
		return ""
	}

	return fragments[index]
}

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

	if len(strings.SplitN(c.AppId, ":", AppIdFragmentCount)) < AppIdFragmentCount {
		return errors.New(fmt.Sprintf("%s does not follow the app id format e.g. 1:123456789:android:xxxxx", c.AppId))
	}

	if c.AccessToken == "" && c.GoogleCredentialsPath == "" {
		logger.Logger.Warn().Msg("we recommend specifying a token or credentials path explicitly")
	} else if c.AccessToken != "" && c.GoogleCredentialsPath != "" {
		logger.Logger.Warn().Msg("the specified access token is prioritized")
	}

	return nil
}

func (c *FirebaseAppDistributionConfig) ProjectNumber() string {
	return AppIdFragment(c.AppId, 1)
}
