package config

import "testing"

func Test_LocalConfig_validateMissingValues(t *testing.T) {
	t.Parallel()

	sampleValue1 := "Sample1"

	cases := map[string]struct {
		config            LocalConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: LocalConfig{
				DestinationPath: sampleValue1,
				AllowOverwrite:  true,
				FileMode:        0644,
				DeleteSource:    true,
			},
			expectedValidness: true,
		},
		"missing-required-fields": {
			config: LocalConfig{
				AllowOverwrite: true,
				FileMode:       0644,
				DeleteSource:   true,
			},
			expectedValidness: false,
		},
		"missing-non-required-fields": {
			config: LocalConfig{
				DestinationPath: sampleValue1,
			},
			expectedValidness: true,
		},
		"zero": {
			config:            LocalConfig{},
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
