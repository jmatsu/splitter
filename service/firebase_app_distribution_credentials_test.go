package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2/google"
)

func Test_credentialsTypeOf(t *testing.T) {
	cases := map[string]struct {
		jsonContent string

		expected    google.CredentialsType
		expectError bool
	}{
		"a service account": {
			jsonContent: `{"type": "service_account", "project_id": "splitter"}`,
			expected:    google.ServiceAccount,
		},
		"an authorized user": {
			jsonContent: `{"type": "authorized_user"}`,
			expected:    google.AuthorizedUser,
		},
		"an external account": {
			jsonContent: `{"type": "external_account"}`,
			expected:    google.ExternalAccount,
		},
		"no type is declared": {
			jsonContent: `{"client_id": "splitter"}`,
			expected:    "",
		},
		"not a json": {
			jsonContent: `not a json`,
			expectError: true,
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := credentialsTypeOf([]byte(c.jsonContent))

			if c.expectError {
				if err == nil {
					t.Error("an error is expected but not raised")
				}
			} else if err != nil {
				t.Errorf("failed to detect the type: %v", err)
			} else if c.expected != actual {
				t.Errorf("the type is expected to be %s but not: %s", c.expected, actual)
			}
		})
	}
}

func Test_FirebaseToken(t *testing.T) {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, ".fixtures", "google_credentials.json")

	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s is not found", path)
		return
	}

	if v, err := FirebaseToken(context.TODO(), path); err != nil {
		t.Errorf("failed to get a token: %v", err)
	} else if !v.Valid() {
		t.Error("the token is invalid")
	}
}
