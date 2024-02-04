package exec

import (
	"context"
	"github.com/pkg/errors"
)

type Altool struct {
	commandLine CommandLine
}

func NewAltool(ctx context.Context) *Altool {
	return &Altool{
		commandLine: NewCommandLine(ctx, nil),
	}
}

func (n *Altool) UploadApp(path, appleID, password string) ([]byte, error) {
	args := []string{
		path,
		"-f", path,
		"-t", "ios",
		"--username", appleID,
		"--password", password,
	}

	stdout, _, err := n.exec("--upload-app", args...)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute altool")
	}

	return stdout, nil
}

func (n *Altool) exec(subcommand string, args ...string) ([]byte, []byte, error) {
	args = append([]string{"altool", subcommand}, args...)
	return n.commandLine.Exec("xcrun", args...)
}
