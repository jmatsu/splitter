package service

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
)

type sideEffect = string

const (
	localCopyOnly         sideEffect = "copied without overwriting"
	localMoveOnly         sideEffect = "moved without overwriting"
	localCopyAndOverwrite sideEffect = "copied and overwrote"
	localMoveAndOverwrite sideEffect = "moved and overwrote"
)

type LocalMoveRequest struct {
	sourceFilePath      string
	destinationFilePath string
	allowOverwrite      bool
	fileMode            os.FileMode
	deleteResource      bool
}

type LocalMoveResponse struct {
	SourceFilePath      string     `json:"source_file_path"`
	DestinationFilePath string     `json:"destination_file_path"`
	SideEffect          sideEffect `json:"side_effect"`
}

func (p *LocalProvider) move(request *LocalMoveRequest) ([]byte, error) {
	sideEffect, err := func() (sideEffect, error) {
		var sideEffect sideEffect

		if _, err := os.Stat(request.sourceFilePath); err != nil {
			return "", errors.New(fmt.Sprintf("%s does not exist", request.sourceFilePath))
		} else if di, err := os.Stat(request.destinationFilePath); err == nil {
			if !request.allowOverwrite {
				return "", errors.New(fmt.Sprintf("%s exists but overwriting is disabled", request.destinationFilePath))
			} else if di.IsDir() {
				return "", errors.New(fmt.Sprintf("directory (%s) as a destination is not supported", request.destinationFilePath))
			}

			if request.deleteResource {
				sideEffect = localMoveAndOverwrite
			} else {
				sideEffect = localCopyAndOverwrite
			}
		} else {
			if request.deleteResource {
				sideEffect = localMoveOnly
			} else {
				sideEffect = localCopyOnly
			}
		}

		var renameFromPath = request.sourceFilePath

		if !request.deleteResource {
			var tmp *os.File

			if v, err := os.CreateTemp("", "local-dest-*"); err != nil {
				return "", errors.Wrap(err, "failed to create a temp file")
			} else {
				tmp = v
				defer tmp.Close()
				renameFromPath = tmp.Name()
			}

			src, err := os.Open(request.sourceFilePath)

			if err != nil {
				return "", errors.Wrapf(err, "failed to open %s", request.sourceFilePath)
			}

			defer src.Close()

			if _, err := io.Copy(tmp, src); err != nil {
				return "", errors.Wrapf(err, "failed to copy %s to %s", request.sourceFilePath, tmp.Name())
			}
		}

		if err := os.Rename(renameFromPath, request.destinationFilePath); err != nil {
			return "", errors.Wrapf(err, "failed to rename %s to %s", renameFromPath, request.destinationFilePath)
		}

		return sideEffect, nil
	}()

	if err != nil {
		return nil, err
	}

	if v, err := os.Stat(request.destinationFilePath); err != nil {
		panic(err)
	} else if v.Mode() == request.fileMode {
		localLogger.Debug().Msgf("%s already has permission %d", request.destinationFilePath, request.fileMode)
	} else if err := os.Chmod(request.destinationFilePath, request.fileMode); err != nil {
		return nil, errors.Wrapf(err, "failed to change file mode of %s to %s", request.destinationFilePath, v.Mode().String())
	} else {
		localLogger.Debug().Msgf("%s has been changed to permission %d", request.destinationFilePath, request.fileMode)
	}

	resp := LocalMoveResponse{
		SourceFilePath:      request.sourceFilePath,
		DestinationFilePath: request.destinationFilePath,
		SideEffect:          sideEffect,
	}

	if bytes, err := json.Marshal(resp); err != nil {
		panic(err)
	} else {
		return bytes, nil
	}
}
