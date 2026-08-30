package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/exec"
)

func Test_TestFlightDeployRequest_NewUploadAppRequest(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config config.TestFlightConfig

		expectedCredential exec.AltoolCredential
	}{
		"with a password": {
			config: config.TestFlightConfig{
				AppleID:  "apple-id",
				Password: "password",
			},
			expectedCredential: exec.AltoolCredential{Password: "password"},
		},
		"with an api key": {
			config: config.TestFlightConfig{
				AppleID:  "apple-id",
				ApiKey:   "api-key",
				IssuerID: "issuer-id",
			},
			expectedCredential: exec.AltoolCredential{ApiKey: "api-key", IssuerID: "issuer-id"},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			provider := NewTestFlightProvider(context.TODO(), &c.config)

			if provider.AppleID != c.config.AppleID {
				t.Errorf("apple id is expected to be %s but %s", c.config.AppleID, provider.AppleID)
			}

			request := TestFlightDeployRequest{
				filePath: "path/to/app.ipa",
				appleID:  c.config.AppleID,
				password: c.config.Password,
				apiKey:   c.config.ApiKey,
				issueID:  c.config.IssuerID,
			}

			uploadRequest := request.NewUploadAppRequest()

			if uploadRequest.filePath != "path/to/app.ipa" {
				t.Errorf("file path is expected to be path/to/app.ipa but %s", uploadRequest.filePath)
			}

			if uploadRequest.appleID != c.config.AppleID {
				t.Errorf("apple id is expected to be %s but %s", c.config.AppleID, uploadRequest.appleID)
			}

			if v := *uploadRequest.NewAltoolCredential(); !reflect.DeepEqual(v, c.expectedCredential) {
				t.Errorf("credential is expected to be %#v but %#v", c.expectedCredential, v)
			}
		})
	}
}

func Test_TestFlightDeployResult(t *testing.T) {
	t.Parallel()

	result := TestFlightDeployResult{RawJson: `{"key":"value"}`}

	if v := result.RawJsonResponse(); v != `{"key":"value"}` {
		t.Errorf("the raw response is expected to be kept but %s", v)
	}

	if _, ok := result.ValueResponse().(TestFlightDeployResult); !ok {
		t.Errorf("the value response is expected to be TestFlightDeployResult but %T", result.ValueResponse())
	}
}
