package ios

import (
	"context"
	"encoding/json"
	"github.com/jmatsu/splitter/internal/util"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type NotaryID = string

type Notarytool struct {
	commandLine util.CommandLine
}

type NotralLogVersion struct {
	LogFormatVersion int `json:"logFormatVersion"`
}

type NotralLogV1 struct {
	NotralLogVersion `json:",inline"`

	JobId          string      `json:"jobId"`
	Status         string      `json:"status"`
	StatusSummary  string      `json:"statusSummary"`
	StatusCode     int         `json:"statusCode"`
	UploadDate     time.Time   `json:"uploadDate"`
	Sha256         string      `json:"sha256"`
	TicketContents interface{} `json:"ticketContents"`
	Issues         []struct {
		Severity     string      `json:"severity"`
		Code         interface{} `json:"code"`
		Path         string      `json:"path"`
		Message      string      `json:"message"`
		DocUrl       interface{} `json:"docUrl"`
		Architecture interface{} `json:"architecture"`
	} `json:"issues"`
}

func NewNotarytool(ctx context.Context) *Notarytool {
	return &Notarytool{
		commandLine: util.NewCommandLine(ctx, nil),
	}
}

func (n *Notarytool) Submit(account *AppleDeveloperAccount, path string, wait bool) (NotaryID, error) {
	args := []string{
		path,
		"--apple-id", account.AppleID,
		"--password", account.Password,
		"--team-ID", account.TeamID,
	}

	if wait {
		args = append(args, "--wait")
	}

	stdout, _, err := n.exec("submit", args...)

	if err != nil {
		return "", errors.New("use `xcron notarytool log <uuid>` to check the details")
	}

	var uuid NotaryID

	for _, line := range strings.Split(string(stdout), "\n") {
		_, s, found := strings.Cut(line, " id: ")

		if !found {
			continue
		}

		uuid, _, _ = strings.Cut(s, " ")
		break
	}

	return uuid, nil
}

func (n *Notarytool) Log(uuid string) (string, error) {
	args := []string{
		uuid,
	}

	stdout, _, err := n.exec("log", args...)

	if err != nil {
		return "", errors.Wrap(err, "failed to get a notary log")
	}

	var version NotralLogVersion

	if err := json.Unmarshal(stdout, &version); err != nil || version.LogFormatVersion != 1 {
		return "", errors.Wrap(err, "notation log format is incompatible")
	}

	var log NotralLogV1

	return uuid, nil
}

func (n *Notarytool) exec(subcommand string, args ...string) ([]byte, []byte, error) {
	args = append([]string{"notarytool", subcommand}, args...)
	return n.commandLine.Exec("xcrun", args...)
}
