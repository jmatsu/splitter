package config

// DeployGateConfig contains the enough values to use DeployGate.
type DeployGateConfig struct {
	ExecutionConfig

	// User#name or Group#name of DeployGate
	AppOwnerName string `json:"app-owner-name,omitempty" env:"DEPLOYGATE_APP_OWNER_NAME" required:"true"`

	// API token of the app owner or who has permission to use their namespace.
	ApiToken string `json:"api-token,omitempty" env:"DEPLOYGATE_API_TOKEN" required:"true"`

	// The existing access key of the distribution
	DistributionAccessKey string `json:"distribution-access-key,omitempty" env:"DEPLOYGATE_DISTRIBUTION_KEY"`

	// A name of a distribution
	DistributionName string `json:"distribution-name,omitempty" env:"DEPLOYGATE_DISTRIBUTION_NAME"`
}

func (c *DeployGateConfig) Validate() error {
	return validateMissingValues(c)
}
