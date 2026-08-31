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
				AccessToken: "AccessToken",
				AppId:       "AppId",
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

func Test_FirebaseAppDistributionConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config            FirebaseAppDistributionConfig
		expectedValidness bool
	}{
		"with an access token": {
			config: FirebaseAppDistributionConfig{
				AppId:       "1:123456:android:xxxxx",
				AccessToken: "AccessToken",
			},
			expectedValidness: true,
		},
		"with a credentials path": {
			config: FirebaseAppDistributionConfig{
				AppId:                 "1:123456:android:xxxxx",
				GoogleCredentialsPath: "path/to/credentials.json",
			},
			expectedValidness: true,
		},
		"with the both of credentials": { // an access token takes priority
			config: FirebaseAppDistributionConfig{
				AppId:                 "1:123456:android:xxxxx",
				AccessToken:           "AccessToken",
				GoogleCredentialsPath: "path/to/credentials.json",
			},
			expectedValidness: true,
		},
		"without any credentials": { // splitter falls back into the application default credentials
			config: FirebaseAppDistributionConfig{
				AppId: "1:123456:android:xxxxx",
			},
			expectedValidness: true,
		},
		"missing app id": {
			config: FirebaseAppDistributionConfig{
				AccessToken: "AccessToken",
			},
			expectedValidness: false,
		},
		"malformed app id": {
			config: FirebaseAppDistributionConfig{
				AppId:       "my-app",
				AccessToken: "AccessToken",
			},
			expectedValidness: false,
		},
		"app id that lacks a uid": {
			config: FirebaseAppDistributionConfig{
				AppId:       "1:123456:android",
				AccessToken: "AccessToken",
			},
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

			if err := c.config.Validate(); (err == nil) != c.expectedValidness {
				t.Errorf("%s case is expected to be %t but %t", name, c.expectedValidness, err == nil)
			}
		})
	}
}

func Test_FirebaseAppDistributionConfig_ProjectNumber(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		appId    string
		expected string
	}{
		"android": {
			appId:    "1:123456789:android:xxxxx",
			expected: "123456789",
		},
		"ios": {
			appId:    "1:987654321:ios:yyyyy",
			expected: "987654321",
		},
		"malformed": { // Validate rejects it beforehand so it must not panic here
			appId:    "my-app",
			expected: "",
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := FirebaseAppDistributionConfig{AppId: c.appId}

			if v := config.ProjectNumber(); v != c.expected {
				t.Errorf("%s case is expected to be %s but %s", name, c.expected, v)
			}
		})
	}
}
