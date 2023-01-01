package config

// DeployGateConfig contains the enough values to use DeployGate.
type DeployGateConfig struct {
	serviceNameHolder `yaml:",inline"`
	ExecutionConfig   `yaml:",inline"`

	// User#name or Organization#name of DeployGate
	AppOwnerName string `yaml:"app-owner-name" env:"DEPLOYGATE_APP_OWNER_NAME" required:"true"`

	// API token of the app owner or who has permission to use their namespace.
	ApiToken string `yaml:"api-token" env:"DEPLOYGATE_API_TOKEN" required:"true"`

	// The existing access key of the distribution
	DistributionAccessKey string `yaml:"distribution-access-key,omitempty" env:"DEPLOYGATE_DISTRIBUTION_KEY"`

	// A name of a distribution
	DistributionName string `yaml:"distribution-name,omitempty" env:"DEPLOYGATE_DISTRIBUTION_NAME"`
}

func (c *DeployGateConfig) Validate() error {
	return validateMissingValues(c)
}
