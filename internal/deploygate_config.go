package internal

type DeployGateConfig struct {
	AppOwnerName string `json:"app-owner-name,omitempty" env:"DEPLOYGATE_APP_OWNER_NAME" required:"true"`
	ApiToken     string `json:"api-token,omitempty" env:"DEPLOYGATE_API_TOKEN" required:"true"`
}
