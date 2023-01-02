package config

type CustomServiceConfig struct {
	serviceNameHolder `yaml:",inline"`
	ExecutionConfig   `yaml:",inline"`

	AuthToken string `yaml:"auth-token"`
}

func (c *CustomServiceConfig) Validate() error {
	return validateMissingValues(c)
}
