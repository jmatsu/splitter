package config

import (
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type serviceNameHolder struct {
	Name string `yaml:"service"`
}

type serviceConfig interface {
	testConfig | DeployGateConfig | LocalConfig | FirebaseAppDistributionConfig | CustomServiceConfig | TestFlightConfig
}

// Set values to the config. Priority: Environment Variables > Given values
func loadServiceConfig[T serviceConfig](v *T, values map[string]interface{}) error {
	if bytes, err := yaml.Marshal(values); err != nil {
		panic(err)
	} else if err := yaml.Unmarshal(bytes, v); err != nil {
		panic(err)
	} else if err := env.Parse(v); err != nil {
		panic(err)
	}

	return nil
}
