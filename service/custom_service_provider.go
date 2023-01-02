package service

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"strings"
)

var customServiceLogger zerolog.Logger

func init() {
	customServiceLogger = logger.Logger.With().Str("service", "custom").Logger()
}

func NewCustomServiceProvider(ctx context.Context, definition *config.CustomServiceDefinition, conf *config.CustomServiceConfig) *CustomServiceProvider {
	scheme, t, _ := strings.Cut(definition.Endpoint, "://")
	hostname, path, _ := strings.Cut(t, "/")

	return &CustomServiceProvider{
		CustomServiceConfig:     *conf,
		CustomServiceDefinition: *definition,
		ctx:                     ctx,
		client:                  net.NewHttpClient(fmt.Sprintf("%s://%s", scheme, hostname)),
		path:                    path,
	}
}

type CustomServiceProvider struct {
	config.CustomServiceConfig
	config.CustomServiceDefinition
	ctx    context.Context
	client *net.HttpClient
	path   string
}

type CustomServiceDeployRequest struct {
	filePath string

	headers map[string][]string
	queries map[string][]string
	form    net.Form
}

func (r *CustomServiceDeployRequest) SetHeader(name string, value string) {
	r.headers[name] = []string{value}
}

func (r *CustomServiceDeployRequest) HasQueryParam(name string) bool {
	_, found := r.queries[name]
	return found
}

func (r *CustomServiceDeployRequest) SetQueryParam(name string, value string) {
	r.queries[name] = []string{value}
}

func (r *CustomServiceDeployRequest) AddQueryParam(name string, value string) {
	r.queries[name] = append(r.queries[name], value)
}

func (r *CustomServiceDeployRequest) SetFormParam(name string, value string) {
	r.form.Set(net.StringField(name, value))
}

func (r *CustomServiceDeployRequest) NewUploadRequest() *CustomServiceUploadAppRequest {
	return &CustomServiceUploadAppRequest{
		filePath: r.filePath,

		headers: r.headers,
		queries: r.queries,
		form:    r.form,
	}
}

type CustomServiceDeployResult struct {
	CustomServiceUploadResponse
}

var _ DeployResult = &CustomServiceDeployResult{}

func (r *CustomServiceDeployResult) RawJsonResponse() string {
	return r.CustomServiceUploadResponse.RawResponse.RawJson()
}

func (r *CustomServiceDeployResult) ValueResponse() any {
	return *r
}

func (p *CustomServiceProvider) Deploy(filePath string, builder func(req *CustomServiceDeployRequest) error) (*CustomServiceDeployResult, error) {
	request := &CustomServiceDeployRequest{
		filePath: filePath,
		headers:  map[string][]string{},
		queries:  map[string][]string{},
	}

	if err := builder(request); err != nil {
		return nil, errors.Wrapf(err, "could not build the request")
	} else {
		customServiceLogger.Debug().Msgf("the request has been built: %v", *request)
	}

	if r, err := p.upload(request.NewUploadRequest()); err != nil {
		return nil, err
	} else {
		return &CustomServiceDeployResult{
			CustomServiceUploadResponse: *r,
		}, nil
	}
}
