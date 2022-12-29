package internal

type DeployGateConfig struct {
	AppOwnerName string `json:"AppOwnerName,omitempty" env:"DEPLOYGATE_APP_OWNER_NAME" required:"true"`
	ApiToken     string `json:"ApiToken,omitempty" env:"DEPLOYGATE_API_TOKEN" required:"true"`
}

func (c DeployGateConfig) validate() error {
	return validateMissingValues(&c)
}
