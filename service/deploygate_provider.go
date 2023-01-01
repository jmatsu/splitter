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

type DeployGateDistributionResult struct {
	deployGateUploadResponse
}

func (r *DeployGateDistributionResult) RawJsonResponse() string {
	return r.deployGateUploadResponse.RawResponse.RawJson()
}

func (r *DeployGateDistributionResult) ValueResponse() any {
	return *r
}

func (p *DeployGateProvider) Distribute(filePath string, builder func(req *DeployGateUploadAppRequest)) (*DeployGateDistributionResult, error) {
	request := NewDeployGateUploadAppRequest(filePath)

	builder(request)

	deployGateLogger.Debug().Msgf("the request has been built: %v", *request)

	if r, err := p.upload(request); err != nil {
		return nil, err
	} else {
		return &DeployGateDistributionResult{
			deployGateUploadResponse: *r,
		}, nil
	}
}
