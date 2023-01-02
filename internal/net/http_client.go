package net

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func NewHttpClient(baseUrl string) *HttpClient {
	baseURL, err := url.ParseRequestURI(baseUrl)

	if err != nil {
		logger.Logger.Err(err).Msgf("%s is invalid", baseUrl)
		return nil
	}

	return &HttpClient{
		client: &http.Client{
			Timeout: config.CurrentConfig().NetworkTimeout(),
		},
		baseURL: *baseURL,
		headers: http.Header{
			"User-Agent": {
				"splitter/", // TODO assign versions
			},
			"Accept": {
				"application/json", // by default
			},
		},
	}
}

type HttpResponse struct {
	Code  int
	bytes []byte
}

type TypedHttpResponse interface {
	Set(r *HttpResponse)
}

func (r *HttpResponse) Successful() bool {
	return 200 <= r.Code && r.Code < 300
}

func (r *HttpResponse) Err() error {
	if r.Successful() {
		return nil
	} else {
		return errors.New(fmt.Sprintf("status = %d, response = %s", r.Code, string(r.bytes)))
	}
}

func (r *HttpResponse) ParseJson(v any) (any, error) {
	if err := json.Unmarshal(r.bytes, v); err != nil {
		return nil, errors.Wrap(err, "failed to parse the response")
	}

	if v, ok := v.(TypedHttpResponse); ok {
		v.Set(r)
	}

	return v, nil
}

func (r *HttpResponse) RawJson() string {
	return string(r.bytes)
}

type HttpClient struct {
	client  *http.Client
	baseURL url.URL
	headers http.Header
}

func (c *HttpClient) WithHeaders(headers http.Header) *HttpClient {
	newClient := c.clone(func(newClient *HttpClient) {
		if headers == nil {
			return
		}

		newClient.setDefaultHeaders(headers)
	})

	return &newClient
}

func (c *HttpClient) setDefaultHeaders(headers http.Header) {
	if headers == nil {
		return
	}

	maps.Copy(c.headers, headers)
}

func (c *HttpClient) DoGet(ctx context.Context, paths []string, queries map[string][]string) (*HttpResponse, error) {
	return c.do(ctx, paths, queries, http.MethodGet, "", nil)
}

func (c *HttpClient) DoPut(ctx context.Context, paths []string, queries map[string][]string, contentType string, requestBody *bytes.Buffer) (*HttpResponse, error) {
	if requestBody != nil {
		return c.do(ctx, paths, queries, http.MethodPut, contentType, requestBody)
	} else {
		return c.do(ctx, paths, queries, http.MethodPut, contentType, nil)
	}
}

func (c *HttpClient) DoPatch(ctx context.Context, paths []string, queries map[string][]string, contentType string, requestBody *bytes.Buffer) (*HttpResponse, error) {
	if requestBody != nil {
		return c.do(ctx, paths, queries, http.MethodPatch, contentType, requestBody)
	} else {
		return c.do(ctx, paths, queries, http.MethodPatch, contentType, nil)
	}
}

func (c *HttpClient) DoPost(ctx context.Context, paths []string, queries map[string][]string, contentType string, requestBody *bytes.Buffer) (*HttpResponse, error) {
	if requestBody != nil {
		return c.do(ctx, paths, queries, http.MethodPost, contentType, requestBody)
	} else {
		return c.do(ctx, paths, queries, http.MethodPost, contentType, nil)
	}
}

func (c *HttpClient) DoPostFileBody(ctx context.Context, paths []string, queries map[string][]string, filePath string) (*HttpResponse, error) {
	if f, err := os.Open(filePath); err != nil {
		return nil, errors.Wrapf(err, "%s is not found", filePath)
	} else if b, err := io.ReadAll(f); err != nil {
		return nil, errors.Wrapf(err, "%s cannot be read", filePath)
	} else {
		buffer := bytes.NewBuffer(b)
		return c.DoPost(ctx, paths, queries, "application/octet-stream", buffer)
	}
}

func (c *HttpClient) DoPostMultipartForm(ctx context.Context, paths []string, queries map[string][]string, form *Form) (*HttpResponse, error) {
	contentType, buffer, err := form.Serialize()

	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize the request form")
	}

	return c.DoPost(ctx, paths, queries, contentType, buffer)
}

func (c *HttpClient) do(ctx context.Context, paths []string, queries map[string][]string, method string, contentType string, requestBody io.Reader) (*HttpResponse, error) {
	if queries == nil {
		queries = map[string][]string{}
	}

	uri := c.baseURL.JoinPath(paths...)
	q := uri.Query()

	for name, values := range queries {
		for _, value := range values {
			if q.Has(name) {
				q.Add(name, value)
			} else {
				q.Set(name, value)
			}
		}
	}

	uri.RawQuery = q.Encode()

	logger.Logger.Debug().Msgf("%s %s", method, uri.String())

	request, err := http.NewRequestWithContext(ctx, method, uri.String(), requestBody)

	if err != nil {
		return nil, errors.Wrap(err, "failed to build the request")
	}

	for name, values := range c.headers {
		request.Header.Set(name, strings.Join(values, ","))
	}

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	resp, err := c.client.Do(request)

	if err != nil {
		return nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if //goland:noinspection GoImportUsedAsName
	bytes, err := io.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		if 200 <= resp.StatusCode && resp.StatusCode < 300 {
			logger.Logger.Trace().Msg(string(bytes))
		}

		return &HttpResponse{
			Code:  resp.StatusCode,
			bytes: bytes,
		}, nil
	}
}

func (c *HttpClient) clone(mapper func(newClient *HttpClient)) HttpClient {
	//goland:noinspection SpellCheckingInspection
	copiee := *c
	mapper(&copiee)
	return copiee
}
