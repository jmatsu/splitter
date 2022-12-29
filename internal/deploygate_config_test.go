package internal

import "testing"

func Test_DeployGateConfig_validate(t *testing.T) {
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

			if err := c.config.validate(); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t", name, c.expectedValidness, err == nil)
			}
		})
	}
}
