package service

import (
	"context"
	"encoding/json"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"strings"
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
	TestFlightUploadAppResponse
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

	var response TestFlightUploadAppResponse

	bytes, err := p.uploadApp(request.NewUploadAppRequest())

	if err != nil {
		return nil, err
	}

	// altool prints nothing on stdout for some of the successful uploads.
	if len(strings.TrimSpace(string(bytes))) > 0 {
		if err := json.Unmarshal(bytes, &response); err != nil {
			return nil, errors.Wrapf(err, "failed to parse the altool output: %s", string(bytes))
		}
	} else {
		bytes = []byte("{}")
	}

	return &TestFlightDeployResult{
		TestFlightUploadAppResponse: response,
		RawJson:                     string(bytes),
	}, nil
}
