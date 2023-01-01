package service

import (
	"context"
	"encoding/json"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"os"
)

var localLogger zerolog.Logger

func init() {
	localLogger = logger.Logger.With().Str("service", "local").Logger()
}

func NewLocalProvider(ctx context.Context, config *config.LocalConfig) *LocalProvider {
	return &LocalProvider{
		LocalConfig: *config,
		ctx:         ctx,
	}
}

type LocalProvider struct {
	config.LocalConfig
	ctx context.Context
}

type LocalDeployRequest struct {
	sourceFilePath      string
	destinationFilePath string
	allowOverwrite      bool
	fileMode            os.FileMode
	deleteResource      bool
}

func (r *LocalDeployRequest) NewMoveRequest() *LocalMoveRequest {
	request := LocalMoveRequest{
		sourceFilePath:      r.sourceFilePath,
		destinationFilePath: r.destinationFilePath,
		allowOverwrite:      r.allowOverwrite,
		fileMode:            r.fileMode,
		deleteResource:      r.deleteResource,
	}

	return &request
}

type LocalDeployResult struct {
	LocalMoveResponse
	RawJson string
}

var _ DeployResult = &LocalDeployResult{}

func (r *LocalDeployResult) RawJsonResponse() string {
	return r.RawJson
}

func (r *LocalDeployResult) ValueResponse() any {
	return *r
}

func (p *LocalProvider) Deploy(filePath string) (*LocalDeployResult, error) {
	request := LocalDeployRequest{
		sourceFilePath:      filePath,
		destinationFilePath: p.DestinationPath,
		allowOverwrite:      p.AllowOverwrite,
		deleteResource:      p.DeleteSource,
	}

	if p.FileMode != 0 {
		request.fileMode = p.FileMode
	} else if v, err := os.Stat(request.sourceFilePath); err == nil { // Do not validate the request here
		request.fileMode = v.Mode()
	}

	var response LocalMoveResponse

	if bytes, err := p.move(request.NewMoveRequest()); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	} else {
		return &LocalDeployResult{
			LocalMoveResponse: response,
			RawJson:           string(bytes),
		}, nil
	}
}
