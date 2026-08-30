package net

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type testEchoResponse struct {
	RequestURI  string              `json:"request_uri"`
	Method      string              `json:"method"`
	ContentType string              `json:"content_type"`
	Body        string              `json:"body"`
	Headers     map[string][]string `json:"headers"`
}

// newEchoServer returns the request as a json response so that tests can assert what the client sent.
func newEchoServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		resp := testEchoResponse{
			RequestURI:  r.RequestURI,
			Method:      r.Method,
			ContentType: r.Header.Get("Content-Type"),
			Body:        string(body),
			Headers:     r.Header,
		}

		if bytes, err := json.Marshal(resp); err != nil {
			t.Errorf("failed to marshal the response: %v", err)
		} else {
			_, _ = fmt.Fprint(w, string(bytes))
		}
	}))

	t.Cleanup(server.Close)

	return server
}

func Test_HttpClient_Do(t *testing.T) {
	t.Parallel()

	server := newEchoServer(t)
	client := NewHttpClient(server.URL)

	cases := map[string]struct {
		do func(ctx context.Context) (*HttpResponse, error)

		expectedMethod      string
		expectedRequestURI  string
		expectedContentType string
		expectedBody        string
	}{
		"get": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoGet(ctx, []string{"path1", "path2"}, nil)
			},
			expectedMethod:     http.MethodGet,
			expectedRequestURI: "/path1/path2",
		},
		"get with queries": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoGet(ctx, []string{"path1"}, map[string][]string{
					"key1": {"value1", "value2"},
					"key2": {"value3"},
				})
			},
			expectedMethod:     http.MethodGet,
			expectedRequestURI: "/path1?key1=value1&key1=value2&key2=value3",
		},
		"put": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoPut(ctx, []string{"path1"}, nil, "application/json", bytes.NewBufferString(`{"key":"value"}`))
			},
			expectedMethod:      http.MethodPut,
			expectedRequestURI:  "/path1",
			expectedContentType: "application/json",
			expectedBody:        `{"key":"value"}`,
		},
		"put without a body": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoPut(ctx, []string{"path1"}, nil, "application/json", nil)
			},
			expectedMethod:      http.MethodPut,
			expectedRequestURI:  "/path1",
			expectedContentType: "application/json",
		},
		"patch": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoPatch(ctx, []string{"path1"}, map[string][]string{
					"updateMask": {"release_notes.text"},
				}, "application/json", bytes.NewBufferString(`{"key":"value"}`))
			},
			expectedMethod:      http.MethodPatch,
			expectedRequestURI:  "/path1?updateMask=release_notes.text",
			expectedContentType: "application/json",
			expectedBody:        `{"key":"value"}`,
		},
		"patch without a body": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoPatch(ctx, []string{"path1"}, nil, "application/json", nil)
			},
			expectedMethod:      http.MethodPatch,
			expectedRequestURI:  "/path1",
			expectedContentType: "application/json",
		},
		"post": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoPost(ctx, []string{"path1"}, nil, "application/json", bytes.NewBufferString(`{"key":"value"}`))
			},
			expectedMethod:      http.MethodPost,
			expectedRequestURI:  "/path1",
			expectedContentType: "application/json",
			expectedBody:        `{"key":"value"}`,
		},
		"post without a body": {
			do: func(ctx context.Context) (*HttpResponse, error) {
				return client.DoPost(ctx, []string{"path1"}, nil, "", nil)
			},
			expectedMethod:     http.MethodPost,
			expectedRequestURI: "/path1",
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp, err := c.do(context.TODO())

			if err != nil {
				t.Fatalf("%s is expected to be success but not: %v", name, err)
			}

			if !resp.Successful() {
				t.Fatalf("%s is expected to be successful but %d", name, resp.Code)
			}

			var echo testEchoResponse

			if _, err := resp.ParseJson(&echo); err != nil {
				t.Fatalf("%s failed to parse the response: %v", name, err)
			}

			if echo.Method != c.expectedMethod {
				t.Errorf("method is expected to be %s but %s", c.expectedMethod, echo.Method)
			}

			if echo.RequestURI != c.expectedRequestURI {
				t.Errorf("request uri is expected to be %s but %s", c.expectedRequestURI, echo.RequestURI)
			}

			if echo.ContentType != c.expectedContentType {
				t.Errorf("content type is expected to be %s but %s", c.expectedContentType, echo.ContentType)
			}

			if echo.Body != c.expectedBody {
				t.Errorf("body is expected to be %s but %s", c.expectedBody, echo.Body)
			}
		})
	}
}

func Test_HttpClient_DoPostFileBody(t *testing.T) {
	t.Parallel()

	server := newEchoServer(t)
	client := NewHttpClient(server.URL)

	content := "sample world"
	path := filepath.Join(t.TempDir(), "file1.txt")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create the testing file: %v", err)
	}

	resp, err := client.DoPostFileBody(context.TODO(), []string{"path1"}, nil, path)

	if err != nil {
		t.Fatalf("failed to post the file body: %v", err)
	}

	var echo testEchoResponse

	if _, err := resp.ParseJson(&echo); err != nil {
		t.Fatalf("failed to parse the response: %v", err)
	}

	if echo.Method != http.MethodPost {
		t.Errorf("method is expected to be %s but %s", http.MethodPost, echo.Method)
	}

	if echo.ContentType != "application/octet-stream" {
		t.Errorf("content type is expected to be application/octet-stream but %s", echo.ContentType)
	}

	if echo.Body != content {
		t.Errorf("body is expected to be %s but %s", content, echo.Body)
	}

	// A missing file must not reach the server.
	if _, err := client.DoPostFileBody(context.TODO(), []string{"path1"}, nil, filepath.Join(t.TempDir(), "not-found")); err == nil {
		t.Errorf("a missing file is expected to be rejected but not")
	}
}

func Test_HttpClient_do_contentTypeHeader(t *testing.T) {
	t.Parallel()

	server := newEchoServer(t)

	cases := map[string]struct {
		headers     http.Header
		contentType string
		expected    string
	}{
		"content type is set unless the header exists": {
			contentType: "application/json",
			expected:    "application/json",
		},
		"the existing header takes priority": {
			headers:     http.Header{"Content-Type": {"application/xml"}},
			contentType: "application/json",
			expected:    "application/xml",
		},
		"multipart form data takes priority over the existing header": {
			headers:     http.Header{"Content-Type": {"application/xml"}},
			contentType: "multipart/form-data; boundary=xxxxx",
			expected:    "multipart/form-data; boundary=xxxxx",
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := NewHttpClient(server.URL).WithHeaders(c.headers)

			resp, err := client.DoPost(context.TODO(), []string{"path1"}, nil, c.contentType, bytes.NewBufferString("body"))

			if err != nil {
				t.Fatalf("failed to post: %v", err)
			}

			var echo testEchoResponse

			if _, err := resp.ParseJson(&echo); err != nil {
				t.Fatalf("failed to parse the response: %v", err)
			}

			if echo.ContentType != c.expected {
				t.Errorf("content type is expected to be %s but %s", c.expected, echo.ContentType)
			}
		})
	}
}

func Test_HttpClient_do_multiValueHeaders(t *testing.T) {
	t.Parallel()

	server := newEchoServer(t)

	client := NewHttpClient(server.URL).WithHeaders(http.Header{
		"X-Multi": {"value1", "value2"},
	})

	resp, err := client.DoGet(context.TODO(), []string{"path1"}, nil)

	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	var echo testEchoResponse

	if _, err := resp.ParseJson(&echo); err != nil {
		t.Fatalf("failed to parse the response: %v", err)
	}

	if expected := []string{"value1", "value2"}; !reflect.DeepEqual(echo.Headers["X-Multi"], expected) {
		t.Errorf("X-Multi is expected to be %v but %v", expected, echo.Headers["X-Multi"])
	}
}

func Test_HttpClient_do_unreachable(t *testing.T) {
	t.Parallel()

	server := newEchoServer(t)
	url := server.URL
	server.Close()

	client := NewHttpClient(url)

	if _, err := client.DoGet(context.TODO(), nil, nil); err == nil {
		t.Errorf("a closed server is expected to be an error but not")
	}
}

func Test_HttpResponse(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		code               int
		bytes              []byte
		expectedSuccessful bool
	}{
		"200":     {code: 200, bytes: []byte(`{"key":"value"}`), expectedSuccessful: true},
		"204":     {code: 204, expectedSuccessful: true},
		"299":     {code: 299, expectedSuccessful: true},
		"300":     {code: 300, bytes: []byte("moved"), expectedSuccessful: false},
		"400":     {code: 400, bytes: []byte("bad request"), expectedSuccessful: false},
		"500":     {code: 500, bytes: []byte("internal server error"), expectedSuccessful: false},
		"invalid": {code: 199, expectedSuccessful: false},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := HttpResponse{Code: c.code, bytes: c.bytes}

			if v := resp.Successful(); v != c.expectedSuccessful {
				t.Errorf("%s is expected to be %t but %t", name, c.expectedSuccessful, v)
			}

			if err := resp.Err(); (err == nil) != c.expectedSuccessful {
				t.Errorf("%s is expected to have an error %t but %t", name, !c.expectedSuccessful, err != nil)
			}

			if v := resp.RawJson(); v != string(c.bytes) {
				t.Errorf("%s is expected to be %s but %s", name, string(c.bytes), v)
			}
		})
	}
}

func Test_HttpResponse_ParseJson_malformed(t *testing.T) {
	t.Parallel()

	resp := HttpResponse{Code: 200, bytes: []byte("not a json")}

	if _, err := resp.ParseJson(&testTypedHttpResponse{}); err == nil {
		t.Errorf("a malformed json is expected to be rejected but not")
	}
}
