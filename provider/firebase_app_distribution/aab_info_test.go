package firebase_app_distribution

import (
	"testing"
)

func Test_checkIntegrationState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		state integrationState

		expected bool
	}{
		"aabIntegrationIntegrated": {
			state:    aabIntegrationIntegrated,
			expected: true,
		},
		"aabIntegrationNonPublished": {
			state:    aabIntegrationNonPublished,
			expected: false,
		},
		"aabIntegrationNotLinked": {
			state:    aabIntegrationNotLinked,
			expected: false,
		},
		"aabIntegrationNoAppFound": {
			state:    aabIntegrationNoAppFound,
			expected: false,
		},
		"aabIntegrationTermsUnaccepted": {
			state:    aabIntegrationTermsUnaccepted,
			expected: false,
		},
		"aabIntegrationUnavailable": {
			state:    aabIntegrationUnavailable,
			expected: false,
		},
		"aabIntegrationUnspecified": {
			state:    aabIntegrationUnspecified,
			expected: false,
		},
		"otherwise": {
			state:    "otherwise",
			expected: true,
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := checkIntegrationState(c.state)

			if c.expected && err == nil || !c.expected && err != nil {
				return
			}

			t.Errorf("%t is expected but not: %v", c.expected, err)
		})
	}
}
