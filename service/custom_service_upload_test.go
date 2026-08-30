package service

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/jmatsu/splitter/internal/config"
)

func Test_CustomServiceProvider_Deploy(t *testing.T) {
	cases := map[string]struct {
		definition config.CustomServiceDefinition
		builder    func(req *CustomServiceDeployRequest) error

		expectedAuthorization string
		expectedQuery         string
		expectedFormValues    map[string][]string
		expectedFormFiles     map[string]string
		expectedBody          string
	}{
		"a token in a header and a source file in a form": {
			definition: config.CustomServiceDefinition{
				SourceFileFormat: config.FormParamsAssignFormatPrefix + "file",
				AuthDefinition: config.CustomAuthDefinition{
					StyleFormat: config.HeadersAssignFormatPrefix + "Authorization",
					ValueFormat: "Bearer %s",
				},
			},
			expectedAuthorization: "Bearer token1",
			expectedFormValues:    map[string][]string{},
			expectedFormFiles:     map[string]string{"file": "app content"},
		},
		"a token in a query and a source file as a request body": {
			definition: config.CustomServiceDefinition{
				SourceFileFormat: config.RequestBodyAssignFormat,
				AuthDefinition: config.CustomAuthDefinition{
					StyleFormat: config.QueryAssignFormatPrefix + "token",
					ValueFormat: "%s",
				},
			},
			expectedQuery: "token=token1",
			expectedBody:  "app content",
		},
		"a token in a form param": {
			definition: config.CustomServiceDefinition{
				SourceFileFormat: config.FormParamsAssignFormatPrefix + "file",
				AuthDefinition: config.CustomAuthDefinition{
					StyleFormat: config.FormParamsAssignFormatPrefix + "token",
					ValueFormat: "token %s",
				},
			},
			expectedFormValues: map[string][]string{"token": {"token token1"}},
			expectedFormFiles:  map[string]string{"file": "app content"},
		},
		"default request values are merged": {
			definition: config.CustomServiceDefinition{
				SourceFileFormat: config.FormParamsAssignFormatPrefix + "file",
				AuthDefinition: config.CustomAuthDefinition{
					StyleFormat: config.HeadersAssignFormatPrefix + "Authorization",
					ValueFormat: "Bearer %s",
				},
				DefaultRequestDefinition: config.DefaultRequestDefinition{
					Headers:    map[string]string{"X-Default": "default-header"},
					Queries:    map[string][]string{"default": {"default-query"}},
					FormParams: map[string]string{"default": "default-form"},
				},
			},
			expectedAuthorization: "Bearer token1",
			expectedQuery:         "default=default-query",
			expectedFormValues:    map[string][]string{"default": {"default-form"}},
			expectedFormFiles:     map[string]string{"file": "app content"},
		},
		"the builder appends values": {
			definition: config.CustomServiceDefinition{
				SourceFileFormat: config.FormParamsAssignFormatPrefix + "file",
				AuthDefinition: config.CustomAuthDefinition{
					StyleFormat: config.HeadersAssignFormatPrefix + "Authorization",
					ValueFormat: "Bearer %s",
				},
			},
			builder: func(req *CustomServiceDeployRequest) error {
				req.SetHeader("X-Custom", "custom-header")
				req.SetFormParam("custom", "custom-form")

				if req.HasQueryParam("custom") {
					return errPreparation
				}

				req.SetQueryParam("custom", "custom-query1")
				req.AddQueryParam("custom", "custom-query2")

				return nil
			},
			expectedAuthorization: "Bearer token1",
			expectedQuery:         "custom=custom-query1&custom=custom-query2",
			expectedFormValues:    map[string][]string{"custom": {"custom-form"}},
			expectedFormFiles:     map[string]string{"file": "app content"},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			server := newTestServer(t, jsonResponder(http.StatusOK, `{"key":"value"}`))

			definition := c.definition
			definition.Endpoint = server.URL + "/path/to/upload"

			provider := NewCustomServiceProvider(context.TODO(), &definition, &config.CustomServiceConfig{
				AuthToken: "token1",
			})

			path := newSourceFile(t, "app.apk", "app content")

			builder := c.builder

			if builder == nil {
				builder = func(req *CustomServiceDeployRequest) error { return nil }
			}

			result, err := provider.Deploy(path, builder)

			if err != nil {
				t.Fatalf("failed to deploy: %v", err)
			}

			request := server.lastRequest(t)

			if request.Method != http.MethodPost {
				t.Errorf("method is expected to be POST but %s", request.Method)
			}

			if expected := "/path/to/upload"; request.Path != expected {
				t.Errorf("path is expected to be %s but %s", expected, request.Path)
			}

			if request.Authorization != c.expectedAuthorization {
				t.Errorf("authorization is expected to be %s but %s", c.expectedAuthorization, request.Authorization)
			}

			if request.RawQuery != c.expectedQuery {
				t.Errorf("query is expected to be %s but %s", c.expectedQuery, request.RawQuery)
			}

			if c.expectedFormValues != nil && !reflect.DeepEqual(request.FormValues, c.expectedFormValues) {
				t.Errorf("form values are expected to be %v but %v", c.expectedFormValues, request.FormValues)
			}

			if c.expectedFormFiles != nil && !reflect.DeepEqual(request.FormFiles, c.expectedFormFiles) {
				t.Errorf("form files are expected to be %v but %v", c.expectedFormFiles, request.FormFiles)
			}

			if request.Body != c.expectedBody {
				t.Errorf("body is expected to be %s but %s", c.expectedBody, request.Body)
			}

			if c.definition.DefaultRequestDefinition.Headers != nil {
				if v := request.Headers.Get("X-Default"); v != "default-header" {
					t.Errorf("the default header is expected to be sent but %s", v)
				}
			}

			if result.RawJsonResponse() != `{"key":"value"}` {
				t.Errorf("the raw response is not kept: %s", result.RawJsonResponse())
			}

			if _, ok := result.ValueResponse().(CustomServiceDeployResult); !ok {
				t.Errorf("the value response is expected to be CustomServiceDeployResult but %T", result.ValueResponse())
			}
		})
	}
}

func Test_CustomServiceProvider_Deploy_failures(t *testing.T) {
	validDefinition := config.CustomServiceDefinition{
		SourceFileFormat: config.FormParamsAssignFormatPrefix + "file",
		AuthDefinition: config.CustomAuthDefinition{
			StyleFormat: config.HeadersAssignFormatPrefix + "Authorization",
			ValueFormat: "Bearer %s",
		},
	}

	cases := map[string]struct {
		definition config.CustomServiceDefinition
		code       int
		builder    func(req *CustomServiceDeployRequest) error
	}{
		"unauthorized": {
			definition: validDefinition,
			code:       http.StatusUnauthorized,
		},
		"server error": {
			definition: validDefinition,
			code:       http.StatusInternalServerError,
		},
		"builder error": {
			definition: validDefinition,
			code:       http.StatusOK,
			builder: func(req *CustomServiceDeployRequest) error {
				return errPreparation
			},
		},
		"a request body is not compatible with form params": {
			definition: config.CustomServiceDefinition{
				SourceFileFormat: config.RequestBodyAssignFormat,
				AuthDefinition: config.CustomAuthDefinition{
					StyleFormat: config.FormParamsAssignFormatPrefix + "token",
					ValueFormat: "%s",
				},
			},
			code: http.StatusOK,
		},
		"a broken auth definition": {
			definition: config.CustomServiceDefinition{
				SourceFileFormat: config.FormParamsAssignFormatPrefix + "file",
				AuthDefinition: config.CustomAuthDefinition{
					StyleFormat: "obababa",
					ValueFormat: "%s",
				},
			},
			code: http.StatusOK,
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			server := newTestServer(t, jsonResponder(c.code, `{}`))

			definition := c.definition
			definition.Endpoint = server.URL + "/path/to/upload"

			provider := NewCustomServiceProvider(context.TODO(), &definition, &config.CustomServiceConfig{
				AuthToken: "token1",
			})

			builder := c.builder

			if builder == nil {
				builder = func(req *CustomServiceDeployRequest) error { return nil }
			}

			path := newSourceFile(t, "app.apk", "app content")

			if _, err := provider.Deploy(path, builder); err == nil {
				t.Errorf("%s case is expected to be failure but not", name)
			}
		})
	}
}
