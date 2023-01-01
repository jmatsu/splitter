package service

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	logger2 "github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/rs/zerolog"
)

var deployGateLogger zerolog.Logger

func init() {
	deployGateLogger = logger2.Logger.With().Str("service", "deploygate").Logger()
}

func NewDeployGateProvider(ctx context.Context, config *config.DeployGateConfig) *DeployGateProvider {
	return &DeployGateProvider{
		DeployGateConfig: *config,
		ctx:              ctx,
		client:           net.NewHttpClient("https://deploygate.com"),
	}
}

type DeployGateProvider struct {
	config.DeployGateConfig
	ctx    context.Context
	client *net.HttpClient
}

type DeployGateDistributionRequest struct {
	filePath            string
	message             string
	distributionOptions deployGateDistributionOptions
	iOSOptions          deployGateIOSOptions
}

func (r *DeployGateDistributionRequest) SetMessage(value string) {
	r.message = value
}

func (r *DeployGateDistributionRequest) SetDistributionAccessKey(value string) {
	r.distributionOptions.AccessKey = value
}

func (r *DeployGateDistributionRequest) SetDistributionName(value string) {
	r.distributionOptions.Name = value
}

func (r *DeployGateDistributionRequest) SetDistributionReleaseNote(value string) {
	r.distributionOptions.ReleaseNote = value
}

func (r *DeployGateDistributionRequest) SetIOSDisableNotification(value bool) {
	r.iOSOptions.DisableNotification = value
}

func (r *DeployGateDistributionRequest) NewUploadRequest() *DeployGateUploadAppRequest {
	request := DeployGateUploadAppRequest{
		filePath:            r.filePath,
		message:             r.message,
		distributionOptions: r.distributionOptions,
		iOSOptions:          r.iOSOptions,
	}

	return &request
}

type DeployGateDistributionResult struct {
	DeployGateUploadResponse
}

var _ DistributionResult = &DeployGateDistributionResult{}

func (r *DeployGateDistributionResult) RawJsonResponse() string {
	return r.DeployGateUploadResponse.RawResponse.RawJson()
}

func (r *DeployGateDistributionResult) ValueResponse() any {
	return *r
}

func (p *DeployGateProvider) Distribute(filePath string, builder func(req *DeployGateDistributionRequest)) (*DeployGateDistributionResult, error) {
	request := NewDeployGateDistributionRequest(filePath)

	builder(request)

	deployGateLogger.Debug().Msgf("the request has been built: %v", *request)

	if r, err := p.upload(request.NewUploadRequest()); err != nil {
		return nil, err
	} else {
		return &DeployGateDistributionResult{
			DeployGateUploadResponse: *r,
		}, nil
	}
}
