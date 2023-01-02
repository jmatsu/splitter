package service

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	logger2 "github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
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

type DeployGateDeployRequest struct {
	filePath            string
	message             string
	distributionOptions deployGateDistributionOptions
	iOSOptions          deployGateIOSOptions
}

func (r *DeployGateDeployRequest) SetMessage(value string) {
	r.message = value
}

func (r *DeployGateDeployRequest) SetDistributionAccessKey(value string) {
	r.distributionOptions.AccessKey = value
}

func (r *DeployGateDeployRequest) SetDistributionName(value string) {
	r.distributionOptions.Name = value
}

func (r *DeployGateDeployRequest) SetDistributionReleaseNote(value string) {
	r.distributionOptions.ReleaseNote = value
}

func (r *DeployGateDeployRequest) SetIOSDisableNotification(value bool) {
	r.iOSOptions.DisableNotification = value
}

func (r *DeployGateDeployRequest) NewUploadRequest() *DeployGateUploadAppRequest {
	request := DeployGateUploadAppRequest{
		filePath:            r.filePath,
		message:             r.message,
		distributionOptions: r.distributionOptions,
		iOSOptions:          r.iOSOptions,
	}

	return &request
}

type DeployGateDeployResult struct {
	DeployGateUploadResponse
}

var _ DeployResult = &DeployGateDeployResult{}

func (r *DeployGateDeployResult) RawJsonResponse() string {
	return r.DeployGateUploadResponse.RawResponse.RawJson()
}

func (r *DeployGateDeployResult) ValueResponse() any {
	return *r
}

func (p *DeployGateProvider) Deploy(filePath string, builder func(req *DeployGateDeployRequest) error) (*DeployGateDeployResult, error) {
	request := &DeployGateDeployRequest{
		filePath: filePath,
	}

	if err := builder(request); err != nil {
		return nil, errors.Wrapf(err, "could not build the request")
	} else {
		deployGateLogger.Debug().Msgf("the request has been built: %v", *request)
	}

	if r, err := p.upload(request.NewUploadRequest()); err != nil {
		return nil, err
	} else {
		return &DeployGateDeployResult{
			DeployGateUploadResponse: *r,
		}, nil
	}
}
