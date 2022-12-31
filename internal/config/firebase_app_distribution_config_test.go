package config

import "testing"

func Test_FirebaseAppDistributionConfig_validateMissingValues(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config            FirebaseAppDistributionConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: FirebaseAppDistributionConfig{
				AccessToken:   "AccessToken",
				ProjectNumber: "ProjectNumber",
			},
			expectedValidness: true,
		},
		"missing-required-fields": { // same to the zero for now but test this explicitly
			config:            FirebaseAppDistributionConfig{},
			expectedValidness: false,
		},
		"zero": {
			config:            FirebaseAppDistributionConfig{},
			expectedValidness: false,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := validateMissingValues(&c.config); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t", name, c.expectedValidness, err == nil)
			}
		})
	}
}
