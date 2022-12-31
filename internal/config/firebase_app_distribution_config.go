package config

import "strings"

type FirebaseAppDistributionConfig struct {
	AccessToken string `json:"access-token,omitempty" required:"true"`
	AppId       string `json:"app-id,omitempty" required:"true"`
}

func (c *FirebaseAppDistributionConfig) Validate() error {
	return validateMissingValues(c)
}

func (c *FirebaseAppDistributionConfig) ProjectNumber() string {
	// <num>:<project number>:<os>:<uid>
	return strings.SplitN(c.AppId, ":", 3)[1]
}
