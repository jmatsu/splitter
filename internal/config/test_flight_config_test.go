package config

import "testing"

func Test_TestFlightConfig_validateMissingValues(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config            TestFlightConfig
		expectedValidness bool
	}{
		"fully-filled-with-password": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				Password: "Password",
			},
			expectedValidness: true,
		},
		"fully-filled-with-api-key": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				ApiKey:   "ApiKey",
				IssuerID: "IssuerID",
			},
			expectedValidness: true,
		},
		"fully-filled-but-both-are-specified": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				ApiKey:   "ApiKey",
				IssuerID: "IssuerID",
				Password: "Password",
			},
			expectedValidness: true,
		},
		"missing-issuer-id-but-api-key-is-found": {
			config: TestFlightConfig{
				AppleID: "AppleID",
				ApiKey:  "ApiKey",
			},
			expectedValidness: true,
		},
		"missing-api-key-but-issuer-id-is-found": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				IssuerID: "IssuerID",
			},
			expectedValidness: true,
		},
		"missing-apple-id": {
			config: TestFlightConfig{
				ApiKey:   "ApiKey",
				IssuerID: "IssuerID",
				Password: "Password",
			},
			expectedValidness: false,
		},
		"zero": {
			config:            TestFlightConfig{},
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

func Test_TestFlightConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config            TestFlightConfig
		expectedValidness bool
	}{
		"fully-filled-with-password": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				Password: "Password",
			},
			expectedValidness: true,
		},
		"fully-filled-with-api-key": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				ApiKey:   "ApiKey",
				IssuerID: "IssuerID",
			},
			expectedValidness: true,
		},
		"fully-filled-but-both-are-specified": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				ApiKey:   "ApiKey",
				IssuerID: "IssuerID",
				Password: "Password",
			},
			expectedValidness: true,
		},
		"missing-issuer-id-but-api-key-is-found": {
			config: TestFlightConfig{
				AppleID: "AppleID",
				ApiKey:  "ApiKey",
			},
			expectedValidness: false,
		},
		"missing-api-key-but-issuer-id-is-found": {
			config: TestFlightConfig{
				AppleID:  "AppleID",
				IssuerID: "IssuerID",
			},
			expectedValidness: false,
		},
		"missing-required-fields": { // same to the zero for now but test this explicitly
			config:            TestFlightConfig{},
			expectedValidness: false,
		},
		"zero": {
			config:            TestFlightConfig{},
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
