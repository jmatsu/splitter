package config

type FirebaseAppDistributionConfig struct {
	AccessToken   string `json:"access-token,omitempty" required:"true"`
	ProjectNumber string `json:"project-number,omitempty" required:"true"`
}

func (c *FirebaseAppDistributionConfig) Validate() error {
	return validateMissingValues(c)
}
