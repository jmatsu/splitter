package exec

import (
	"context"
	"github.com/pkg/errors"
)

type Altool struct {
	commandLine CommandLine
}

type AltoolCredential struct {
	Password string
	IssuerID string
	ApiKey   string
}

func NewAltool(ctx context.Context) *Altool {
	return &Altool{
		commandLine: NewCommandLine(ctx, nil),
	}
}

func (n *Altool) UploadApp(path, appleID string, credential *AltoolCredential) ([]byte, error) {
	args := []string{
		"-f", path,
		"-t", "ios",
		"--username", appleID,
	}

	if credential.Password != "" {
		args = append(args, "--password", credential.Password)
	} else {
		args = append(args, "--apiKey", credential.ApiKey, "--apiIssuer", credential.IssuerID)
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
