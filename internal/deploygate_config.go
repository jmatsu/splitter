package internal

type DeployGateConfig struct {
	AppOwnerName *string `json:"AppOwnerName,omitempty" env:"DEPLOYGATE_APP_OWNER_NAME"`
	ApiToken *string `json:"ApiToken,omitempty" env:"DEPLOYGATE_API_TOKEN"`
}
