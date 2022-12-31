package net

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	client = http.Client{
		Timeout: 10 * time.Minute,
	}
}

var client http.Client

func GetHttpClient(baseUrl string) *HttpClient {
	baseURL, err := url.ParseRequestURI(baseUrl)

	if err != nil {
		logger.Logger.Err(err).Msgf("%s is invalid", baseUrl)
		return nil
	}

	return &HttpClient{
		client:  client,
		baseURL: baseURL,
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
	client  http.Client
	baseURL *url.URL
	headers http.Header
}

func (c *HttpClient) SetDefaultHeaders(headers http.Header) {
	if headers == nil {
		return
	}

	maps.Copy(c.headers, headers)
}

func (c *HttpClient) WithHeaders(headers http.Header) *HttpClient {
	newClient := c.clone(func(newClient *HttpClient) {
		if headers == nil {
			return
		}

		newClient.SetDefaultHeaders(headers)
	})

	return &newClient
}

func (c *HttpClient) DoPostMultipartForm(ctx context.Context, paths []string, form *Form) (int, []byte, error) {
	contentType, buffer, err := form.Serialize()

	if err != nil {
		return 0, nil, fmt.Errorf("failed to serialize the request form: %v", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.JoinPath(paths...).String(), buffer)

	if err != nil {
		return 0, nil, fmt.Errorf("failed to build the request: %v", err)
	}

	for name, value := range c.headers {
		request.Header.Set(name, value[0])
	}

	request.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(request)

	if err != nil {
		return 0, nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if bytes, err := io.ReadAll(resp.Body); err != nil {
		return 0, nil, err
	} else {
		return resp.StatusCode, bytes, nil
	}
}

func (c *HttpClient) clone(mapper func(newClient *HttpClient)) HttpClient {
	//goland:noinspection SpellCheckingInspection
	copiee := *c
	mapper(&copiee)
	return copiee
}
