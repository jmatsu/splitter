package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	internalHttp "github.com/jmatsu/splitter/internal/http"
)

func Test_GetHttpClient(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		baseUrl         string
		expectedSuccess bool
	}{
		"localhost with port": {
			baseUrl:         "http://localhost:3000",
			expectedSuccess: true,
		},
		"http with port": {
			baseUrl:         "http://example.com:3000",
			expectedSuccess: true,
		},
		"http without port": {
			baseUrl:         "http://example.com",
			expectedSuccess: true,
		},
		"https with port": {
			baseUrl:         "https://example.com:8080",
			expectedSuccess: true,
		},
		"https without port": {
			baseUrl:         "https://example.com",
			expectedSuccess: true,
		},
		"without scheme": {
			baseUrl:         "example.com",
			expectedSuccess: false,
		},
		"non http URL": {
			baseUrl:         "example",
			expectedSuccess: false,
		},
		"zero": {
			expectedSuccess: false,
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := GetHttpClient(c.baseUrl)
			actual := client != nil

			if actual == c.expectedSuccess {
				return
			}

			t.Errorf("%s case is expected to be %t but %t", name, c.expectedSuccess, actual)
		})
	}
}

func Test_HttpClient_SetDefaultHeaders(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		defaultHeaders map[string][]string
	}{
		"override presets": {
			defaultHeaders: map[string][]string{
				"User-Agent": {"sample"},
			},
		},
		"present": {
			defaultHeaders: map[string][]string{
				"TestHeader1": {"sample"},
			},
		},
		"empty": {
			defaultHeaders: map[string][]string{},
		},
		"zero": {},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := GetHttpClient("https://example.com")
			client.SetDefaultHeaders(c.defaultHeaders)

			if c.defaultHeaders != nil {
				for key, value := range c.defaultHeaders {
					if !reflect.DeepEqual(client.headers[key], value) {
						t.Errorf("%s case is expected to be assigned but not: %s = %v", name, key, value)
					}
				}
			}
		})
	}
}

func Test_HttpClient_WithHeaders(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		newHeaders map[string][]string
	}{
		"override presets": {
			newHeaders: map[string][]string{
				"User-Agent": {"sample"},
			},
		},
		"present": {
			newHeaders: map[string][]string{
				"TestHeader1": {"sample"},
			},
		},
		"empty": {
			newHeaders: map[string][]string{},
		},
		"zero": {},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := GetHttpClient("https://example.com")
			newClient := client.WithHeaders(c.newHeaders)

			if client == newClient {
				t.Errorf("clone requires a different instance but not")
				return
			}

			if c.newHeaders != nil {
				for key, value := range c.newHeaders {
					if !reflect.DeepEqual(newClient.headers[key], value) {
						t.Errorf("%s case is expected to be assigned but not: %s = %v", name, key, value)
					}
				}
			}
		})
	}
}

type testResponse struct {
	RequestURI  string
	Fields      map[string]string
	Method      string
	ContentType string
}

func Test_HttpClient_DoPostMultipartForm(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp(os.TempDir(), "splitter")

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	if err != nil {
		panic(err)
	}

	var (
		testFilePath    = filepath.Join(tempDir, "file1.txt")
		testFileContent = "sample world"
	)

	if f, err := os.Create(testFilePath); err != nil {
		t.Errorf("failed to create the testing file: %v", err)
	} else if _, err := f.WriteString(testFileContent); err != nil {
		t.Errorf("failed to write the content of %s", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(32 << 20)

		fields := map[string]string{}

		for name := range r.Form {
			fields[name] = r.Form.Get(name)
		}

		for name := range r.PostForm {
			fields[name] = r.PostForm.Get(name)
		}

		if r.MultipartForm != nil {
			for name := range r.MultipartForm.Value {
				if values := r.MultipartForm.Value[name]; len(values) > 0 {
					fields[name] = values[0]
				} else {
					fields[name] = "no values are found."
				}
			}

			for name := range r.MultipartForm.File {
				func() {
					if values := r.MultipartForm.File[name]; len(values) > 0 {
						if f, err := values[0].Open(); err != nil {
							fields[name] = err.Error()
						} else {
							defer f.Close()
							bytes, _ := io.ReadAll(f)
							fields[name] = string(bytes)
						}
					} else {
						fields[name] = "no values are found."
					}
				}()
			}
		}

		contentType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";")

		resp := testResponse{
			RequestURI:  r.RequestURI,
			Method:      r.Method,
			ContentType: contentType,
			Fields:      fields,
		}

		if bytes, err := json.Marshal(resp); err != nil {
			t.Errorf("failed to marshal the response: %v", err)
		} else {
			_, _ = fmt.Fprintln(w, string(bytes))
		}
	}))

	defer server.Close()

	cases := map[string]struct {
		paths    []string
		form     internalHttp.Form
		expected testResponse
	}{
		"filled": {
			paths: []string{"path1", "path2"},
			form: internalHttp.Form{
				Fields: []internalHttp.ValueField{
					internalHttp.StringField("param1", "value1"),
					internalHttp.FileField("file1", testFilePath),
				},
			},
			expected: testResponse{
				RequestURI: "/path1/path2",
				Fields: map[string]string{
					"param1": "value1",
					"file1":  testFileContent,
				},
				Method:      "POST",
				ContentType: "multipart/form-data",
			},
		},
		"zero": {
			expected: testResponse{
				RequestURI:  "/",
				Method:      "POST",
				Fields:      map[string]string{},
				ContentType: "multipart/form-data",
			},
		},
	}

	client := GetHttpClient(server.URL)

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()

			var resp testResponse

			if _, bytes, err := client.DoPostMultipartForm(ctx, c.paths, &c.form); err != nil {
				t.Errorf("%s is expected to be success but not: %v", name, err)
			} else if err := json.Unmarshal(bytes, &resp); err != nil {
				t.Errorf("%s is expected to be success but not due to comparison failure: %s", name, string(bytes))
			}

			if !reflect.DeepEqual(c.expected, resp) {
				t.Errorf("%s is expected to be equal but not: %v", name, resp)
			}
		})
	}
}

func Test_HttpClient_clone(t *testing.T) {
	client := GetHttpClient("https://example.com")

	newClient := client.clone(func(newClient *HttpClient) {
		if client == newClient {
			t.Errorf("clone requires a different instance in mapper but not")
		}
	})

	if client == &newClient {
		t.Errorf("clone requires a different instance but not")
	}
}
