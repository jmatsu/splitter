package internal

import "testing"

type testConfig struct {
	ValueParam           string  `json:"param1" env:"TEST_PARAM1"`
	PointerParam         *string `json:"param2" env:"TEST_PARAM2"`
	RequiredValueParam   string  `json:"param3" env:"TEST_PARAM3" required:"true"`
	RequiredPointerParam *string `json:"param4" env:"TEST_PARAM4" required:"true"`
}

func (c testConfig) validate() error {
	return validateMissingValues(&c)
}

func Test_validateMissingValues(t *testing.T) {
	t.Parallel()

	sampleValue1 := "Sample1"
	sampleValue2 := "Sample2"

	cases := map[string]struct {
		config            testConfig
		expectedValidness bool
	}{
		"fully-filled": {
			config: testConfig{
				ValueParam:           sampleValue1,
				PointerParam:         &sampleValue2,
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: true,
		},
		"pointer-filled": {
			config: testConfig{
				PointerParam:         &sampleValue2,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: false,
		},
		"pointer-non-filled": {
			config: testConfig{
				ValueParam:         sampleValue1,
				RequiredValueParam: sampleValue1,
			},
			expectedValidness: false,
		},
		"required-values-filled": {
			config: testConfig{
				RequiredValueParam:   sampleValue1,
				RequiredPointerParam: &sampleValue2,
			},
			expectedValidness: true,
		},
		"non-required-values-filled": {
			config: testConfig{
				ValueParam:   sampleValue1,
				PointerParam: &sampleValue2,
			},
			expectedValidness: false,
		},
		"zero": {
			config:            testConfig{},
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
