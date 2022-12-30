package config

type DeployGateConfig struct {
	AppOwnerName          string `json:"app-owner-name,omitempty" env:"DEPLOYGATE_APP_OWNER_NAME" required:"true"`
	ApiToken              string `json:"api-token,omitempty" env:"DEPLOYGATE_API_TOKEN" required:"true"`
	DistributionAccessKey string `json:"distribution-access-key,omitempty" env:"DEPLOYGATE_DISTRIBUTION_KEY"`
	DistributionName      string `json:"distribution-name,omitempty" env:"DEPLOYGATE_DISTRIBUTION_NAME"`
}
