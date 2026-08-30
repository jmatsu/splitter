package service

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/net"
)

const deployGateUploadResponseBody = `{
  "results": {
    "os_name": "android",
    "name": "Example",
    "package_name": "io.github.jmatsu.splitter.example",
    "revision": 12,
    "version_code": "3",
    "version_name": "1.0.2",
    "sdk_version": 21,
    "raw_sdk_version": "21",
    "target_sdk_version": 33,
    "file": "https://deploygate.com/path/to/binary",
    "user": { "name": "owner1" },
    "distribution": {
      "access_key": "dist_key",
      "title": "dist_title",
      "release_note": "note",
      "url": "https://deploygate.com/distributions/dist_key"
    }
  }
}`

// newTestDeployGateProvider points a provider at the test server instead of deploygate.com.
func newTestDeployGateProvider(t *testing.T, serverURL string, conf config.DeployGateConfig) *DeployGateProvider {
	t.Helper()

	provider := NewDeployGateProvider(context.TODO(), &conf)
	provider.client = net.NewHttpClient(serverURL)

	return provider
}

func Test_DeployGateProvider_Deploy(t *testing.T) {
	server := newTestServer(t, jsonResponder(http.StatusOK, deployGateUploadResponseBody))

	provider := newTestDeployGateProvider(t, server.URL, config.DeployGateConfig{
		AppOwnerName: "owner1",
		ApiToken:     "token1",
	})

	path := newSourceFile(t, "app.apk", "app content")

	result, err := provider.Deploy(path, func(req *DeployGateDeployRequest) error {
		req.SetMessage("message1")
		req.SetDistributionName("dist_name")
		req.SetDistributionReleaseNote("note1")
		req.SetIOSDisableNotification(true)
		return nil
	})

	if err != nil {
		t.Fatalf("failed to deploy: %v", err)
	}

	request := server.lastRequest(t)

	if request.Method != http.MethodPost {
		t.Errorf("method is expected to be POST but %s", request.Method)
	}

	if expected := "/api/users/owner1/apps"; request.Path != expected {
		t.Errorf("path is expected to be %s but %s", expected, request.Path)
	}

	if expected := "Bearer token1"; request.Authorization != expected {
		t.Errorf("authorization is expected to be %s but %s", expected, request.Authorization)
	}

	if v := request.FormFiles["file"]; v != "app content" {
		t.Errorf("the source file is expected to be uploaded but %s", v)
	}

	expectedForm := map[string][]string{
		"message":           {"message1"},
		"distribution_name": {"dist_name"},
		"release_note":      {"note1"},
		"disable_notify":    {"true"},
	}

	if !reflect.DeepEqual(request.FormValues, expectedForm) {
		t.Errorf("form is expected to be %v but %v", expectedForm, request.FormValues)
	}

	if v := result.Results.User.Name; v != "owner1" {
		t.Errorf("owner name is expected to be owner1 but %s", v)
	}

	if v := result.Results.Revision; v != 12 {
		t.Errorf("revision is expected to be 12 but %d", v)
	}

	if result.Results.Distribution == nil {
		t.Fatal("distribution is expected to be parsed but nil")
	}

	if v := result.Results.Distribution.AccessKey; v != "dist_key" {
		t.Errorf("distribution access key is expected to be dist_key but %s", v)
	}

	if result.RawJsonResponse() != deployGateUploadResponseBody {
		t.Errorf("the raw response is not kept: %s", result.RawJsonResponse())
	}

	if _, ok := result.ValueResponse().(DeployGateDeployResult); !ok {
		t.Errorf("the value response is expected to be DeployGateDeployResult but %T", result.ValueResponse())
	}
}

func Test_DeployGateProvider_Deploy_withDistributionAccessKey(t *testing.T) {
	server := newTestServer(t, jsonResponder(http.StatusOK, deployGateUploadResponseBody))

	provider := newTestDeployGateProvider(t, server.URL, config.DeployGateConfig{
		AppOwnerName: "owner1",
		ApiToken:     "token1",
	})

	path := newSourceFile(t, "app.apk", "app content")

	if _, err := provider.Deploy(path, func(req *DeployGateDeployRequest) error {
		req.SetDistributionAccessKey("dist_key")
		return nil
	}); err != nil {
		t.Fatalf("failed to deploy: %v", err)
	}

	request := server.lastRequest(t)

	if v := request.FormValues["distribution_key"]; !reflect.DeepEqual(v, []string{"dist_key"}) {
		t.Errorf("distribution key is expected to be sent but %v", v)
	}

	if _, found := request.FormValues["distribution_name"]; found {
		t.Errorf("distribution name is not expected to be sent")
	}
}

func Test_DeployGateProvider_Deploy_failures(t *testing.T) {
	cases := map[string]struct {
		code    int
		body    string
		builder func(req *DeployGateDeployRequest) error
	}{
		"unauthorized": {
			code: http.StatusUnauthorized,
			body: `{"error": true, "message": "invalid token"}`,
		},
		"server error": {
			code: http.StatusInternalServerError,
			body: `{"error": true}`,
		},
		"malformed response": {
			code: http.StatusOK,
			body: `not a json`,
		},
		"builder error": {
			code: http.StatusOK,
			body: deployGateUploadResponseBody,
			builder: func(req *DeployGateDeployRequest) error {
				return errPreparation
			},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			server := newTestServer(t, jsonResponder(c.code, c.body))

			provider := newTestDeployGateProvider(t, server.URL, config.DeployGateConfig{
				AppOwnerName: "owner1",
				ApiToken:     "token1",
			})

			builder := c.builder

			if builder == nil {
				builder = func(req *DeployGateDeployRequest) error { return nil }
			}

			path := newSourceFile(t, "app.apk", "app content")

			if _, err := provider.Deploy(path, builder); err == nil {
				t.Errorf("%s case is expected to be failure but not", name)
			}
		})
	}
}

func Test_DeployGateDeployRequest_NewUploadRequest(t *testing.T) {
	t.Parallel()

	request := DeployGateDeployRequest{filePath: "path/to/app.apk"}

	request.SetMessage("message1")
	request.SetDistributionAccessKey("dist_key")
	request.SetDistributionName("dist_name")
	request.SetDistributionReleaseNote("note1")
	request.SetIOSDisableNotification(true)

	expected := DeployGateUploadAppRequest{
		filePath: "path/to/app.apk",
		message:  "message1",
		distributionOptions: deployGateDistributionOptions{
			Name:        "dist_name",
			AccessKey:   "dist_key",
			ReleaseNote: "note1",
		},
		iOSOptions: deployGateIOSOptions{DisableNotification: true},
	}

	if v := *request.NewUploadRequest(); !reflect.DeepEqual(v, expected) {
		t.Errorf("the upload request is expected to be %#v but %#v", expected, v)
	}
}
