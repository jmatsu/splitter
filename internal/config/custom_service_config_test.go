package config

import "testing"

func Test_CustomServiceConfig_validateMissingValues(t *testing.T) {
	t.Parallel()

	sampleValue1 := "Sample1"

	cases := map[string]struct {
		config            CustomServiceConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: CustomServiceConfig{
				AuthToken: sampleValue1,
			},
			expectedValidness: true,
		},
		"zero": {
			config:            CustomServiceConfig{},
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
