package local

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	logger2 "github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"os"
)

var logger zerolog.Logger

func init() {
	logger = logger2.Logger.With().Str("provider", "local").Logger()
}

type Provider struct {
	config.LocalConfig
	ctx context.Context
}

func NewProvider(ctx context.Context, config *config.LocalConfig) *Provider {
	return &Provider{
		LocalConfig: *config,
		ctx:         ctx,
	}
}

type MoveRequest struct {
	SourceFilePath      string
	DestinationFilePath string
	AllowOverride       bool
	FileMode            os.FileMode
	DeleteResource      bool
}

func (p *Provider) Distribute(filePath string) (*DistributionResult, error) {
	request := MoveRequest{
		SourceFilePath:      filePath,
		DestinationFilePath: p.DestinationPath,
		AllowOverride:       p.AllowOverwrite,
		DeleteResource:      p.DeleteSource,
	}

	if p.FileMode != 0 {
		request.FileMode = p.FileMode
	} else if v, err := os.Stat(request.SourceFilePath); err == nil { // Do not validate the request here
		request.FileMode = v.Mode()
	}

	var response moveResponse

	if bytes, err := p.distribute(&request); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	} else {
		return &DistributionResult{
			moveResponse: response,
			RawJson:      string(bytes),
		}, nil
	}
}

func (p *Provider) distribute(request *MoveRequest) ([]byte, error) {
	sideEffect, err := func() (sideEffect, error) {
		var sideEffect sideEffect

		if _, err := os.Stat(request.SourceFilePath); err != nil {
			return "", errors.New(fmt.Sprintf("%s does not exist", request.SourceFilePath))
		} else if di, err := os.Stat(request.DestinationFilePath); err == nil {
			if !request.AllowOverride {
				return "", errors.New(fmt.Sprintf("%s exists but overwriting is disabled", request.DestinationFilePath))
			} else if di.IsDir() {
				return "", errors.New(fmt.Sprintf("directory (%s) as a destination is not supported", request.DestinationFilePath))
			}

			if request.DeleteResource {
				sideEffect = moveAndOverride
			} else {
				sideEffect = copyAndOverride
			}
		} else {
			if request.DeleteResource {
				sideEffect = moveOnly
			} else {
				sideEffect = copyOnly
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
		logger.Debug().Msgf("%s already has permission %d", request.DestinationFilePath, request.FileMode)
	} else if err := os.Chmod(request.DestinationFilePath, request.FileMode); err != nil {
		return nil, errors.Wrapf(err, "failed to change file mode of %s to %s", request.DestinationFilePath, v.Mode().String())
	} else {
		logger.Debug().Msgf("%s has been changed to permission %d", request.DestinationFilePath, request.FileMode)
	}

	resp := moveResponse{
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
