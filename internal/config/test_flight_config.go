package config

import "time"

type TestFlightConfig struct {
	KeyId               string `yaml:"key-id"`
	IssuerId            string `yaml:"issuer-id"`
	TokenExpiryDuration string `yaml:"token-expiry-duration"`
	KeyPath             string `yaml:"key-path"`
}

func (c *TestFlightConfig) TokenExpiry() time.Duration {
	if d, err := time.ParseDuration(c.TokenExpiryDuration); err != nil {
		panic(err)
	} else if d > 20*time.Minute {
		panic(err)
	} else {
		return d
	}
}
