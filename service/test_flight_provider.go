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

type testFlightDeployRequest struct {
	appleID  string
	password string
	filePath string
}

func (r *testFlightDeployRequest) NewUploadAppRequest() *testFlightUploadAppRequest {
	request := testFlightUploadAppRequest{
		appleID:  r.appleID,
		password: r.password,
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

func (p *TestFlightProvider) Deploy(filePath string) (*TestFlightDeployResult, error) {
	request := testFlightDeployRequest{
		filePath: filePath,
		appleID:  p.AppleID,
		password: p.Password,
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
