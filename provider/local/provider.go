package local

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	logger2 "github.com/jmatsu/splitter/internal/logger"
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

func (p *Provider) Distribute(filePath string) ([]byte, error) {
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

	return p.distribute(&request)
}

func (p *Provider) distribute(request *MoveRequest) ([]byte, error) {
	sideEffect, err := func() (sideEffect, error) {
		var sideEffect sideEffect

		if _, err := os.Stat(request.SourceFilePath); err != nil {
			return "", fmt.Errorf("%s does not exist", request.SourceFilePath)
		} else if di, err := os.Stat(request.DestinationFilePath); err == nil {
			if !request.AllowOverride {
				return "", fmt.Errorf("%s exists but overwriting is disabled", request.DestinationFilePath)
			} else if di.IsDir() {
				return "", fmt.Errorf("directory (%s) as a destination is not supported", request.DestinationFilePath)
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
				return "", fmt.Errorf("failed to create a temp file: %v", err)
			} else {
				tmp = v
				defer tmp.Close()
				renameFromPath = tmp.Name()
			}

			src, err := os.Open(request.SourceFilePath)

			if err != nil {
				return "", fmt.Errorf("failed to open %s: %v", request.SourceFilePath, err)
			}

			defer src.Close()

			if _, err := io.Copy(tmp, src); err != nil {
				return "", fmt.Errorf("failed to copy %s to %s: %v", request.SourceFilePath, tmp.Name(), err)
			}
		}

		if err := os.Rename(renameFromPath, request.DestinationFilePath); err != nil {
			return "", fmt.Errorf("failed to rename %s to %s: %v", renameFromPath, request.DestinationFilePath, err)
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
		return nil, fmt.Errorf("failed to change file mode of %s to %s: %v", request.DestinationFilePath, v.Mode().String(), err)
	} else {
		logger.Debug().Msgf("%s has been changed to permission %d", request.DestinationFilePath, request.FileMode)
	}

	resp := MoveResponse{
		SourceFilePath:      request.SourceFilePath,
		DestinationFilePath: request.DestinationFilePath,
		SideEffect:          sideEffect,
	}

	if bytes, err := json.Marshal(resp); err != nil {
		return nil, err
	} else {
		return bytes, nil
	}
}
