package ios

import (
	"context"
	"github.com/jmatsu/splitter/internal/util"
	"github.com/pkg/errors"
)

type Ditto struct {
	commandLine util.CommandLine
}

func NewDitto(ctx context.Context) *Ditto {
	return &Ditto{
		commandLine: util.NewCommandLine(ctx, nil),
	}
}

func (d *Ditto) CreateZip(sourcePath string, zipPath string) error {
	args := []string{
		"-c",
		"-k",
		"--keepParent",
		sourcePath,
		zipPath,
	}

	_, _, err := d.commandLine.Exec("ditto", args...)

	if err != nil {
		return errors.New("ditto failed: failed to create a zip file")
	}

	return nil
}