package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/net"
)

const (
	testFirebaseAppId         = "1:123456789:android:xxxxx"
	testFirebaseProjectNumber = "123456789"
	testFirebaseOperation     = "projects/123456789/apps/1:123456789:android:xxxxx/releases/-/operations/op1"
	testFirebaseRelease       = "projects/123456789/apps/1:123456789:android:xxxxx/releases/release1"
)

// newTestFirebaseAppDistributionProvider points a provider at the test server instead of googleapis.com.
func newTestFirebaseAppDistributionProvider(t *testing.T, serverURL string, conf config.FirebaseAppDistributionConfig) *FirebaseAppDistributionProvider {
	t.Helper()

	provider := NewFirebaseAppDistributionProvider(context.TODO(), &conf)
	provider.client = net.NewHttpClient(serverURL)

	return provider
}

// firebaseHandler emulates the App Distribution endpoints splitter talks to.
type firebaseHandler struct {
	integrationState appBundleIntegrationState
	aabInfoCode      int
	operationDone    bool
	releaseNote      string
	distributeCode   int
}

func (h *firebaseHandler) handle(r recordedRequest, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case strings.HasSuffix(r.Path, "/aabInfo"):
		code := h.aabInfoCode

		if code == 0 {
			code = http.StatusOK
		}

		state := h.integrationState

		if state == "" {
			state = aabIntegrationIntegrated
		}

		w.WriteHeader(code)
		_, _ = fmt.Fprintf(w, `{"integrationState": %q, "testCertificate": {"hashSha1": "sha1"}}`, state)
	case strings.HasSuffix(r.Path, "releases:upload"):
		_, _ = fmt.Fprintf(w, `{"name": %q}`, testFirebaseOperation)
	case strings.HasSuffix(r.Path, ":distribute"):
		code := h.distributeCode

		if code == 0 {
			code = http.StatusOK
		}

		w.WriteHeader(code)
		_, _ = fmt.Fprint(w, `{}`)
	case r.Method == http.MethodPatch:
		var request firebaseAppDistributionUpdateReleaseRequest

		_ = json.Unmarshal([]byte(r.Body), &request)

		h.releaseNote = request.ReleaseNote.Text

		_, _ = fmt.Fprintf(w, `{"name": %q, "displayVersion": "1.0.2", "buildVersion": "3", "releaseNotes": {"text": %q}}`, testFirebaseRelease, request.ReleaseNote.Text)
	default: // the operation state
		_, _ = fmt.Fprintf(w, `{"name": %q, "done": %t, "response": {"result": "RELEASE_CREATED", "release": {"name": %q, "displayVersion": "1.0.2", "buildVersion": "3"}}}`, testFirebaseOperation, h.operationDone, testFirebaseRelease)
	}
}

func Test_FirebaseAppDistributionProvider_Deploy(t *testing.T) {
	handler := &firebaseHandler{operationDone: true}
	server := newTestServer(t, handler.handle)

	provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
		AppId:        testFirebaseAppId,
		AccessToken:  "token1",
		GroupAliases: []string{"group1"},
	})

	path := newSourceFile(t, "app.apk", "app content")

	result, err := provider.Deploy(path, func(req *FirebaseAppDistributionDeployRequest) error {
		req.SetReleaseNote("note1")
		req.SetTesterEmails([]string{"tester@example.com"})
		return nil
	})

	if err != nil {
		t.Fatalf("failed to deploy: %v", err)
	}

	if handler.releaseNote != "note1" {
		t.Errorf("the release note is expected to be updated to note1 but %s", handler.releaseNote)
	}

	if !reflect.DeepEqual(result.GroupAliases, []string{"group1"}) {
		t.Errorf("group aliases are expected to be %v but %v", []string{"group1"}, result.GroupAliases)
	}

	if !reflect.DeepEqual(result.TesterEmails, []string{"tester@example.com"}) {
		t.Errorf("tester emails are expected to be %v but %v", []string{"tester@example.com"}, result.TesterEmails)
	}

	if result.AabInfo == nil || !result.AabInfo.Available() {
		t.Errorf("aab info is expected to be available but %#v", result.AabInfo)
	}

	if v := result.Response.Release.ReleaseNote; v == nil || v.Text != "note1" {
		t.Errorf("the release note is expected to be kept in the result but %#v", v)
	}

	if result.RawJsonResponse() == "" {
		t.Errorf("the raw response is expected to be kept but empty")
	}

	if _, ok := result.ValueResponse().(FirebaseAppDistributionDeployResult); !ok {
		t.Errorf("the value response is expected to be FirebaseAppDistributionDeployResult but %T", result.ValueResponse())
	}

	// Every request must be authorized and hit the documented endpoints.
	var uploaded, distributed bool

	for _, request := range server.requests {
		if request.Authorization != "Bearer token1" {
			t.Errorf("%s is expected to be authorized but %s", request.Path, request.Authorization)
		}

		switch {
		case strings.HasSuffix(request.Path, "releases:upload"):
			uploaded = true

			if expected := fmt.Sprintf("/upload/v1/projects/%s/apps/%s/releases:upload", testFirebaseProjectNumber, testFirebaseAppId); request.Path != expected {
				t.Errorf("the upload path is expected to be %s but %s", expected, request.Path)
			}

			if request.Body != "app content" {
				t.Errorf("the source file is expected to be a request body but %s", request.Body)
			}

			if v := request.Headers.Get("X-Goog-Upload-File-Name"); v != "app.apk" {
				t.Errorf("the upload file name is expected to be app.apk but %s", v)
			}

			if v := request.Headers.Get("X-Goog-Upload-Protocol"); v != "raw" {
				t.Errorf("the upload protocol is expected to be raw but %s", v)
			}
		case strings.HasSuffix(request.Path, ":distribute"):
			distributed = true

			if expected := fmt.Sprintf("/v1/%s:distribute", testFirebaseRelease); request.Path != expected {
				t.Errorf("the distribute path is expected to be %s but %s", expected, request.Path)
			}
		case request.Method == http.MethodPatch:
			if expected := "updateMask=release_notes.text"; request.RawQuery != expected {
				t.Errorf("the update mask is expected to be %s but %s", expected, request.RawQuery)
			}
		}
	}

	if !uploaded {
		t.Errorf("the app is expected to be uploaded but not")
	}

	if !distributed {
		t.Errorf("the release is expected to be distributed but not")
	}
}

func Test_FirebaseAppDistributionProvider_Deploy_appBundle(t *testing.T) {
	cases := map[string]struct {
		integrationState appBundleIntegrationState
		expectedSuccess  bool
	}{
		"integrated":                 {integrationState: aabIntegrationIntegrated, expectedSuccess: true},
		"play account is not linked": {integrationState: aabIntegrationNotLinked},
		"app is not published":       {integrationState: aabIntegrationNonPublished},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			handler := &firebaseHandler{operationDone: true, integrationState: c.integrationState}
			server := newTestServer(t, handler.handle)

			provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
				AppId:       testFirebaseAppId,
				AccessToken: "token1",
			})

			path := newSourceFile(t, "app.aab", "app content")

			_, err := provider.Deploy(path, func(req *FirebaseAppDistributionDeployRequest) error { return nil })

			if (err == nil) != c.expectedSuccess {
				t.Errorf("%s case is expected to be %t but %t: %v", name, c.expectedSuccess, err == nil, err)
			}
		})
	}
}

func Test_FirebaseAppDistributionProvider_Deploy_unavailableAabInfo(t *testing.T) {
	cases := map[string]struct {
		fileName        string
		expectedSuccess bool
	}{
		// An app bundle cannot be uploaded without knowing the play store integration state.
		"app bundle": {fileName: "app.aab"},
		// An apk does not need it so the deployment must go on.
		"apk": {fileName: "app.apk", expectedSuccess: true},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			handler := &firebaseHandler{operationDone: true, aabInfoCode: http.StatusForbidden}
			server := newTestServer(t, handler.handle)

			provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
				AppId:       testFirebaseAppId,
				AccessToken: "token1",
			})

			path := newSourceFile(t, c.fileName, "app content")

			result, err := provider.Deploy(path, func(req *FirebaseAppDistributionDeployRequest) error { return nil })

			if (err == nil) != c.expectedSuccess {
				t.Fatalf("%s case is expected to be %t but %t: %v", name, c.expectedSuccess, err == nil, err)
			}

			if c.expectedSuccess && result.AabInfo != nil {
				t.Errorf("aab info is expected to be empty but %#v", result.AabInfo)
			}
		})
	}
}

func Test_FirebaseAppDistributionProvider_Deploy_failures(t *testing.T) {
	t.Run("builder error", func(t *testing.T) {
		handler := &firebaseHandler{operationDone: true}
		server := newTestServer(t, handler.handle)

		provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
			AppId:       testFirebaseAppId,
			AccessToken: "token1",
		})

		path := newSourceFile(t, "app.apk", "app content")

		if _, err := provider.Deploy(path, func(req *FirebaseAppDistributionDeployRequest) error {
			return errPreparation
		}); err == nil {
			t.Errorf("a builder error is expected to be propagated but not")
		}
	})

	t.Run("upload error", func(t *testing.T) {
		server := newTestServer(t, jsonResponder(http.StatusForbidden, `{"error": {"message": "permission denied"}}`))

		provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
			AppId:       testFirebaseAppId,
			AccessToken: "token1",
		})

		path := newSourceFile(t, "app.apk", "app content")

		if _, err := provider.Deploy(path, func(req *FirebaseAppDistributionDeployRequest) error { return nil }); err == nil {
			t.Errorf("an upload error is expected to be propagated but not")
		}
	})

	t.Run("missing credentials", func(t *testing.T) {
		server := newTestServer(t, jsonResponder(http.StatusOK, `{}`))

		provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
			AppId:                 testFirebaseAppId,
			GoogleCredentialsPath: "path/to/not-found.json",
		})

		path := newSourceFile(t, "app.apk", "app content")

		if _, err := provider.Deploy(path, func(req *FirebaseAppDistributionDeployRequest) error { return nil }); err == nil {
			t.Errorf("a missing credentials file is expected to be an error but not")
		}
	})
}

func Test_FirebaseAppDistributionProvider_waitForOperationDone_timeout(t *testing.T) {
	original := config.CurrentConfig().WaitTimeout()

	config.SetGlobalWaitTimeout("1s")
	t.Cleanup(func() {
		config.SetGlobalWaitTimeout(original.String())
	})

	handler := &firebaseHandler{operationDone: false}
	server := newTestServer(t, handler.handle)

	provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
		AppId:       testFirebaseAppId,
		AccessToken: "token1",
	})

	if _, err := provider.waitForOperationDone(&firebaseAppDistributionGetOperationStateRequest{
		operationName: testFirebaseOperation,
	}); err == nil {
		t.Errorf("an unfinished operation is expected to time out but not")
	}
}

func Test_FirebaseAppDistributionProvider_getOperationState(t *testing.T) {
	cases := map[string]struct {
		code            int
		body            string
		expectedSuccess bool
	}{
		"done":               {code: http.StatusOK, body: `{"name": "op1", "done": true}`, expectedSuccess: true},
		"not found":          {code: http.StatusNotFound, body: `{"error": {}}`},
		"malformed response": {code: http.StatusOK, body: `not a json`},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			server := newTestServer(t, jsonResponder(c.code, c.body))

			provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
				AppId:       testFirebaseAppId,
				AccessToken: "token1",
			})

			resp, err := provider.getOperationState(&firebaseAppDistributionGetOperationStateRequest{
				operationName: testFirebaseOperation,
			})

			if (err == nil) != c.expectedSuccess {
				t.Fatalf("%s case is expected to be %t but %t: %v", name, c.expectedSuccess, err == nil, err)
			}

			if !c.expectedSuccess {
				return
			}

			if expected := fmt.Sprintf("/v1/%s", testFirebaseOperation); server.lastRequest(t).Path != expected {
				t.Errorf("path is expected to be %s but %s", expected, server.lastRequest(t).Path)
			}

			if !resp.Done {
				t.Errorf("the operation is expected to be done but not")
			}
		})
	}
}

func Test_FirebaseAppDistributionProvider_getAabInfo(t *testing.T) {
	cases := map[string]struct {
		code              int
		body              string
		expectedSuccess   bool
		expectedAvailable bool
	}{
		"integrated": {
			code:              http.StatusOK,
			body:              fmt.Sprintf(`{"integrationState": %q}`, aabIntegrationIntegrated),
			expectedSuccess:   true,
			expectedAvailable: true,
		},
		"not linked": {
			code:            http.StatusOK,
			body:            fmt.Sprintf(`{"integrationState": %q}`, aabIntegrationNotLinked),
			expectedSuccess: true,
		},
		"forbidden": {
			code: http.StatusForbidden,
			body: `{"error": {}}`,
		},
		"malformed response": {
			code: http.StatusOK,
			body: `not a json`,
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			server := newTestServer(t, jsonResponder(c.code, c.body))

			provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
				AppId:       testFirebaseAppId,
				AccessToken: "token1",
			})

			resp, err := provider.getAabInfo(&firebaseAppDistributionAabInfoRequest{
				projectNumber: testFirebaseProjectNumber,
				appId:         testFirebaseAppId,
			})

			if (err == nil) != c.expectedSuccess {
				t.Fatalf("%s case is expected to be %t but %t: %v", name, c.expectedSuccess, err == nil, err)
			}

			if !c.expectedSuccess {
				return
			}

			if expected := fmt.Sprintf("/v1/projects/%s/apps/%s/aabInfo", testFirebaseProjectNumber, testFirebaseAppId); server.lastRequest(t).Path != expected {
				t.Errorf("path is expected to be %s but %s", expected, server.lastRequest(t).Path)
			}

			if v := resp.Available(); v != c.expectedAvailable {
				t.Errorf("availability is expected to be %t but %t", c.expectedAvailable, v)
			}
		})
	}
}

func Test_FirebaseAppDistributionProvider_updateReleaseNote(t *testing.T) {
	cases := map[string]struct {
		code            int
		body            string
		expectedSuccess bool
	}{
		"updated":            {code: http.StatusOK, body: `{"name": "release1", "releaseNotes": {"text": "note1"}}`, expectedSuccess: true},
		"forbidden":          {code: http.StatusForbidden, body: `{"error": {}}`},
		"malformed response": {code: http.StatusOK, body: `not a json`},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			server := newTestServer(t, jsonResponder(c.code, c.body))

			provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
				AppId:       testFirebaseAppId,
				AccessToken: "token1",
			})

			release := FirebaseAppDistributionReleaseFragment{Name: testFirebaseRelease}

			resp, err := provider.updateReleaseNote(release.NewUpdateRequest("note1"))

			if (err == nil) != c.expectedSuccess {
				t.Fatalf("%s case is expected to be %t but %t: %v", name, c.expectedSuccess, err == nil, err)
			}

			if !c.expectedSuccess {
				return
			}

			request := server.lastRequest(t)

			if request.Method != http.MethodPatch {
				t.Errorf("method is expected to be PATCH but %s", request.Method)
			}

			if expected := `{"name":"` + testFirebaseRelease + `","releaseNotes":{"text":"note1"}}`; request.Body != expected {
				t.Errorf("body is expected to be %s but %s", expected, request.Body)
			}

			if resp.ReleaseNote == nil || resp.ReleaseNote.Text != "note1" {
				t.Errorf("the release note is expected to be parsed but %#v", resp.ReleaseNote)
			}
		})
	}
}

func Test_FirebaseAppDistributionProvider_distributeRelease(t *testing.T) {
	cases := map[string]struct {
		code            int
		expectedSuccess bool
	}{
		"distributed": {code: http.StatusOK, expectedSuccess: true},
		"forbidden":   {code: http.StatusForbidden},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			server := newTestServer(t, jsonResponder(c.code, `{}`))

			provider := newTestFirebaseAppDistributionProvider(t, server.URL, config.FirebaseAppDistributionConfig{
				AppId:       testFirebaseAppId,
				AccessToken: "token1",
			})

			release := FirebaseAppDistributionReleaseFragment{Name: testFirebaseRelease}

			err := provider.distributeRelease(release.NewDistributeRequest([]string{"tester@example.com"}, []string{"group1"}))

			if (err == nil) != c.expectedSuccess {
				t.Fatalf("%s case is expected to be %t but %t: %v", name, c.expectedSuccess, err == nil, err)
			}

			if !c.expectedSuccess {
				return
			}

			request := server.lastRequest(t)

			if expected := `{"testerEmails":["tester@example.com"],"groupAliases":["group1"]}`; request.Body != expected {
				t.Errorf("body is expected to be %s but %s", expected, request.Body)
			}
		})
	}
}

func Test_FirebaseAppDistributionDeployRequest(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		appId    string
		filePath string

		expectedOsName   string
		expectedFileType string
	}{
		"android apk": {
			appId:            "1:123456789:android:xxxxx",
			filePath:         "path/to/app.apk",
			expectedOsName:   "android",
			expectedFileType: "apk",
		},
		"android aab": {
			appId:            "1:123456789:android:xxxxx",
			filePath:         "path/to/app.aab",
			expectedOsName:   "android",
			expectedFileType: "aab",
		},
		"ios ipa": {
			appId:            "1:123456789:ios:yyyyy",
			filePath:         "path/to/app.IPA",
			expectedOsName:   "ios",
			expectedFileType: "ipa",
		},
		"no extension": {
			appId:          "1:123456789:android:xxxxx",
			filePath:       "path/to/app",
			expectedOsName: "android",
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := FirebaseAppDistributionDeployRequest{
				projectNumber: testFirebaseProjectNumber,
				appId:         c.appId,
				filePath:      c.filePath,
			}

			if v := request.OsName(); v != c.expectedOsName {
				t.Errorf("os name is expected to be %s but %s", c.expectedOsName, v)
			}

			if v := request.fileType(); v != c.expectedFileType {
				t.Errorf("file type is expected to be %s but %s", c.expectedFileType, v)
			}

			expected := FirebaseAppDistributionUploadAppRequest{
				projectNumber: testFirebaseProjectNumber,
				appId:         c.appId,
				filePath:      c.filePath,
			}

			if v := *request.NewUploadRequest(); !reflect.DeepEqual(v, expected) {
				t.Errorf("the upload request is expected to be %#v but %#v", expected, v)
			}
		})
	}
}
