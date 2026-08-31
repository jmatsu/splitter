package config

type CustomServiceConfig struct {
	serviceNameHolder `yaml:",inline"`

	AuthToken string `yaml:"auth-token" required:"true"`
}

func (c *CustomServiceConfig) Validate() error {
	return validateMissingValues(c)
}
