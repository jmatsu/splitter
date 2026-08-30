package service

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// recordedRequest holds what a provider actually sent to a service.
type recordedRequest struct {
	Method        string
	Path          string
	RawQuery      string
	Authorization string
	Headers       http.Header
	Body          string
	FormValues    map[string][]string
	FormFiles     map[string]string
}

// testServer replies with the given status and body, and records the last request.
type testServer struct {
	*httptest.Server

	requests []recordedRequest
}

func (s *testServer) lastRequest(t *testing.T) recordedRequest {
	t.Helper()

	if len(s.requests) == 0 {
		t.Fatal("no request has been recorded")
	}

	return s.requests[len(s.requests)-1]
}

// newTestServer builds a server that responds by the given handler and records every request.
func newTestServer(t *testing.T, handler func(r recordedRequest, w http.ResponseWriter)) *testServer {
	t.Helper()

	server := &testServer{}

	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded := recordedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			RawQuery:      r.URL.RawQuery,
			Authorization: r.Header.Get("Authorization"),
			Headers:       r.Header.Clone(),
			FormValues:    map[string][]string{},
			FormFiles:     map[string]string{},
		}

		if err := r.ParseMultipartForm(32 << 20); err == nil && r.MultipartForm != nil {
			for name, values := range r.MultipartForm.Value {
				recorded.FormValues[name] = values
			}

			for name, headers := range r.MultipartForm.File {
				if len(headers) == 0 {
					continue
				}

				if f, err := headers[0].Open(); err == nil {
					//goland:noinspection GoUnhandledErrorResult
					defer f.Close()

					bytes := make([]byte, headers[0].Size)
					_, _ = f.Read(bytes)
					recorded.FormFiles[name] = string(bytes)
				}
			}
		} else if bytes, err := io.ReadAll(r.Body); err == nil {
			recorded.Body = string(bytes)
		}

		server.requests = append(server.requests, recorded)

		handler(recorded, w)
	}))

	t.Cleanup(server.Close)

	return server
}

// jsonResponder replies with the given status code and body.
func jsonResponder(code int, body string) func(r recordedRequest, w http.ResponseWriter) {
	return func(_ recordedRequest, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	}
}

// newSourceFile creates an app file to be uploaded.
func newSourceFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create %s: %v", path, err)
	}

	return path
}

// errPreparation is a sentinel error to fail a request builder.
var errPreparation = errors.New("failed to prepare the request")
