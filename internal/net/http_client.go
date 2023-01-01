package net

import (
	"bytes"
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"net/url"
	"os"
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

func (c *HttpClient) DoGet(ctx context.Context, paths []string, queries map[string]string) (int, []byte, error) {
	return c.do(ctx, paths, queries, http.MethodGet, "", nil)
}

func (c *HttpClient) DoPut(ctx context.Context, paths []string, queries map[string]string, contentType string, requestBody *bytes.Buffer) (int, []byte, error) {
	if requestBody != nil {
		return c.do(ctx, paths, queries, http.MethodPut, contentType, requestBody)
	} else {
		return c.do(ctx, paths, queries, http.MethodPut, contentType, nil)
	}
}

func (c *HttpClient) DoPatch(ctx context.Context, paths []string, queries map[string]string, contentType string, requestBody *bytes.Buffer) (int, []byte, error) {
	if requestBody != nil {
		return c.do(ctx, paths, queries, http.MethodPatch, contentType, requestBody)
	} else {
		return c.do(ctx, paths, queries, http.MethodPatch, contentType, nil)
	}
}

func (c *HttpClient) DoPost(ctx context.Context, paths []string, queries map[string]string, contentType string, requestBody *bytes.Buffer) (int, []byte, error) {
	if requestBody != nil {
		return c.do(ctx, paths, queries, http.MethodPost, contentType, requestBody)
	} else {
		return c.do(ctx, paths, queries, http.MethodPost, contentType, nil)
	}
}

func (c *HttpClient) DoPostFileBody(ctx context.Context, paths []string, filePath string) (int, []byte, error) {
	if f, err := os.Open(filePath); err != nil {
		return 0, nil, errors.Wrapf(err, "%s is not found", filePath)
	} else if b, err := io.ReadAll(f); err != nil {
		return 0, nil, errors.Wrapf(err, "%s cannot be read", filePath)
	} else {
		buffer := bytes.NewBuffer(b)
		return c.DoPost(ctx, paths, nil, "application/octet-stream", buffer)
	}
}

func (c *HttpClient) DoPostMultipartForm(ctx context.Context, paths []string, form *Form) (int, []byte, error) {
	contentType, buffer, err := form.Serialize()

	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to serialize the request form")
	}

	return c.DoPost(ctx, paths, nil, contentType, buffer)
}

func (c *HttpClient) do(ctx context.Context, paths []string, queries map[string]string, method string, contentType string, requestBody io.Reader) (int, []byte, error) {
	if queries == nil {
		queries = map[string]string{}
	}

	uri := c.baseURL.JoinPath(paths...)
	q := uri.Query()

	for name, value := range queries {
		q.Set(name, value)
	}

	uri.RawQuery = q.Encode()

	logger.Logger.Debug().Msgf("%s %s", method, uri.String())

	request, err := http.NewRequestWithContext(ctx, method, uri.String(), requestBody)

	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to build the request")
	}

	for name, value := range c.headers {
		request.Header.Set(name, value[0])
	}

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	resp, err := c.client.Do(request)

	if err != nil {
		return 0, nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if //goland:noinspection GoImportUsedAsName
	bytes, err := io.ReadAll(resp.Body); err != nil {
		return 0, nil, err
	} else {
		if 200 <= resp.StatusCode && resp.StatusCode < 300 {
			logger.Logger.Trace().Msg(string(bytes))
		}

		return resp.StatusCode, bytes, nil
	}
}

func (c *HttpClient) clone(mapper func(newClient *HttpClient)) HttpClient {
	//goland:noinspection SpellCheckingInspection
	copiee := *c
	mapper(&copiee)
	return copiee
}
