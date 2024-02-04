package service

import (
	"context"
	"encoding/json"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var testFlightLogger zerolog.Logger

func init() {
	testFlightLogger = logger.Logger.With().Str("service", "test flight").Logger()
}

func NewTestFlightProvider(ctx context.Context, config *config.TestFlightConfig) *TestFlightProvider {
	return &TestFlightProvider{
		TestFlightConfig: *config,
		ctx:              ctx,
	}
}

type TestFlightProvider struct {
	config.TestFlightConfig
	ctx context.Context
}

type TestFlightDeployRequest struct {
	appleID  string
	password string
	issueID  string
	apiKey   string
	filePath string
}

func (r *TestFlightDeployRequest) NewUploadAppRequest() *TestFlightUploadAppRequest {
	request := TestFlightUploadAppRequest{
		appleID:  r.appleID,
		password: r.password,
		issuerID: r.issueID,
		apiKey:   r.apiKey,
		filePath: r.filePath,
	}

	return &request
}

type TestFlightDeployResult struct {
	testFlightUploadAppResponse
	RawJson string
}

var _ DeployResult = &TestFlightDeployResult{}

func (r *TestFlightDeployResult) RawJsonResponse() string {
	return r.RawJson
}

func (r *TestFlightDeployResult) ValueResponse() any {
	return *r
}

func (p *TestFlightProvider) Deploy(filePath string, builder func(req *TestFlightDeployRequest) error) (*TestFlightDeployResult, error) {
	request := &TestFlightDeployRequest{
		filePath: filePath,
		appleID:  p.AppleID,
		password: p.Password,
		issueID:  p.IssuerID,
		apiKey:   p.ApiKey,
	}

	if err := builder(request); err != nil {
		return nil, errors.Wrapf(err, "could not build the request")
	} else {
		testFlightLogger.Debug().Msgf("the request has been built: %v", *request)
	}

	var response testFlightUploadAppResponse

	if bytes, err := p.uploadApp(request.NewUploadAppRequest()); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	} else {
		return &TestFlightDeployResult{
			testFlightUploadAppResponse: response,
			RawJson:                     string(bytes),
		}, nil
	}
}
