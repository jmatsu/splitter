package config

import "testing"

func Test_DeployGateConfig_validateMissingValues(t *testing.T) {
	t.Parallel()

	sampleValue1 := "Sample1"
	sampleValue2 := "Sample2"

	cases := map[string]struct {
		config            DeployGateConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: DeployGateConfig{
				AppOwnerName: sampleValue1,
				ApiToken:     sampleValue2,
			},
			expectedValidness: true,
		},
		"required-values-filled": {
			config: DeployGateConfig{
				AppOwnerName: sampleValue1,
				ApiToken:     sampleValue2,
			},
			expectedValidness: true,
		},
		"missing-app-owner-name": {
			config: DeployGateConfig{
				ApiToken: sampleValue2,
			},
			expectedValidness: false,
		},
		"missing-api-token": {
			config: DeployGateConfig{
				AppOwnerName: sampleValue1,
			},
			expectedValidness: false,
		},
		"zero": {
			config:            DeployGateConfig{},
			expectedValidness: false,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := validateMissingValues(&c.config); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expectedServices to be %t but %t", name, c.expectedValidness, err == nil)
			}
		})
	}
}

func Test_DeployGateConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config            DeployGateConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: DeployGateConfig{
				AppOwnerName:          "AppOwnerName",
				ApiToken:              "ApiToken",
				DistributionAccessKey: "DistributionAccessKey",
				DistributionName:      "DistributionName",
			},
			expectedValidness: true,
		},
		"required only": {
			config: DeployGateConfig{
				AppOwnerName: "AppOwnerName",
				ApiToken:     "ApiToken",
			},
			expectedValidness: true,
		},
		"missing api token": {
			config: DeployGateConfig{
				AppOwnerName: "AppOwnerName",
			},
			expectedValidness: false,
		},
		"missing app owner name": {
			config: DeployGateConfig{
				ApiToken: "ApiToken",
			},
			expectedValidness: false,
		},
		"zero": {
			config:            DeployGateConfig{},
			expectedValidness: false,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := c.config.Validate(); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t", name, c.expectedValidness, err == nil)
			}
		})
	}
}
