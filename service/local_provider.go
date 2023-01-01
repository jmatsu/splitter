package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"os"
)

var localLogger zerolog.Logger

type sideEffect = string

const (
	localCopyOnly         sideEffect = "copied without overwriting"
	localMoveOnly         sideEffect = "moved without overwriting"
	localCopyAndOverwrite sideEffect = "copied and overwrote"
	localMoveAndOverwrite sideEffect = "moved and overwrote"
)

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

type LocalDistributionResult struct {
	LocalMoveResponse
	RawJson string
}

func (r *LocalDistributionResult) RawJsonResponse() string {
	return r.RawJson
}

func (r *LocalDistributionResult) ValueResponse() any {
	return *r
}

type LocalMoveRequest struct {
	SourceFilePath      string
	DestinationFilePath string
	AllowOverwrite      bool
	FileMode            os.FileMode
	DeleteResource      bool
}

type LocalMoveResponse struct {
	SourceFilePath      string     `json:"source_file_path"`
	DestinationFilePath string     `json:"destination_file_path"`
	SideEffect          sideEffect `json:"side_effect"`
}

func (p *LocalProvider) Distribute(filePath string) (*LocalDistributionResult, error) {
	request := LocalMoveRequest{
		SourceFilePath:      filePath,
		DestinationFilePath: p.DestinationPath,
		AllowOverwrite:      p.AllowOverwrite,
		DeleteResource:      p.DeleteSource,
	}

	if p.FileMode != 0 {
		request.FileMode = p.FileMode
	} else if v, err := os.Stat(request.SourceFilePath); err == nil { // Do not validate the request here
		request.FileMode = v.Mode()
	}

	var response LocalMoveResponse

	if bytes, err := p.distribute(&request); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	} else {
		return &LocalDistributionResult{
			LocalMoveResponse: response,
			RawJson:           string(bytes),
		}, nil
	}
}

func (p *LocalProvider) distribute(request *LocalMoveRequest) ([]byte, error) {
	sideEffect, err := func() (sideEffect, error) {
		var sideEffect sideEffect

		if _, err := os.Stat(request.SourceFilePath); err != nil {
			return "", errors.New(fmt.Sprintf("%s does not exist", request.SourceFilePath))
		} else if di, err := os.Stat(request.DestinationFilePath); err == nil {
			if !request.AllowOverwrite {
				return "", errors.New(fmt.Sprintf("%s exists but overwriting is disabled", request.DestinationFilePath))
			} else if di.IsDir() {
				return "", errors.New(fmt.Sprintf("directory (%s) as a destination is not supported", request.DestinationFilePath))
			}

			if request.DeleteResource {
				sideEffect = localMoveAndOverwrite
			} else {
				sideEffect = localCopyAndOverwrite
			}
		} else {
			if request.DeleteResource {
				sideEffect = localMoveOnly
			} else {
				sideEffect = localCopyOnly
			}
		}

		var renameFromPath = request.SourceFilePath

		if !request.DeleteResource {
			var tmp *os.File

			if v, err := os.CreateTemp("", "local-dest-*"); err != nil {
				return "", errors.Wrap(err, "failed to create a temp file")
			} else {
				tmp = v
				defer tmp.Close()
				renameFromPath = tmp.Name()
			}

			src, err := os.Open(request.SourceFilePath)

			if err != nil {
				return "", errors.Wrapf(err, "failed to open %s", request.SourceFilePath)
			}

			defer src.Close()

			if _, err := io.Copy(tmp, src); err != nil {
				return "", errors.Wrapf(err, "failed to copy %s to %s", request.SourceFilePath, tmp.Name())
			}
		}

		if err := os.Rename(renameFromPath, request.DestinationFilePath); err != nil {
			return "", errors.Wrapf(err, "failed to rename %s to %s", renameFromPath, request.DestinationFilePath)
		}

		return sideEffect, nil
	}()

	if err != nil {
		return nil, err
	}

	if v, err := os.Stat(request.DestinationFilePath); err != nil {
		panic(err)
	} else if v.Mode() == request.FileMode {
		localLogger.Debug().Msgf("%s already has permission %d", request.DestinationFilePath, request.FileMode)
	} else if err := os.Chmod(request.DestinationFilePath, request.FileMode); err != nil {
		return nil, errors.Wrapf(err, "failed to change file mode of %s to %s", request.DestinationFilePath, v.Mode().String())
	} else {
		localLogger.Debug().Msgf("%s has been changed to permission %d", request.DestinationFilePath, request.FileMode)
	}

	resp := LocalMoveResponse{
		SourceFilePath:      request.SourceFilePath,
		DestinationFilePath: request.DestinationFilePath,
		SideEffect:          sideEffect,
	}

	if bytes, err := json.Marshal(resp); err != nil {
		panic(err)
	} else {
		return bytes, nil
	}
}
