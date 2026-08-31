package task

import (
	"encoding/json"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Dumper writes a deployment result to <dist dir>/<run name>/<name>.json so that the following
// commands in a user's sequence can consume it without parsing the stdout.
type Dumper struct {
	// serviceName is a name of the service that a app has been deployed to.
	serviceName string

	// deploymentName is empty unless a deployment in a config file is used.
	deploymentName string

	sourceFilePath string
}

func NewDumper(serviceName string, deploymentName string, sourceFilePath string) *Dumper {
	return &Dumper{
		serviceName:    serviceName,
		deploymentName: deploymentName,
		sourceFilePath: sourceFilePath,
	}
}

type dumpedResult struct {
	Service        string  `json:"service"`
	Deployment     *string `json:"deployment"`
	RunName        string  `json:"run_name"`
	DeployedAt     string  `json:"deployed_at"`
	SourceFilePath string  `json:"source_file_path"`

	service.NormalizedResult

	Values any             `json:"values"`
	Raw    json.RawMessage `json:"raw"`
}

func (d *Dumper) Dump(r service.DeployResult) error {
	conf := config.CurrentConfig()

	dir := filepath.Join(conf.DistDir(), conf.RunName())

	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create %s", dir)
	}

	result := dumpedResult{
		Service:          d.serviceName,
		RunName:          conf.RunName(),
		DeployedAt:       time.Now().UTC().Format(time.RFC3339),
		SourceFilePath:   d.sourceFilePath,
		NormalizedResult: r.NormalizedResponse(),
		Values:           r.ValueResponse(),
		Raw:              rawMessage(r.RawJsonResponse()),
	}

	if d.deploymentName != "" {
		result.Deployment = &d.deploymentName
	}

	bytes, err := json.MarshalIndent(result, "", "  ")

	if err != nil {
		return errors.Wrap(err, "failed to normalize the response")
	}

	path := filepath.Join(dir, fileName(d.name()))

	if err := os.WriteFile(path, append(bytes, '\n'), 0644); err != nil {
		return errors.Wrapf(err, "failed to dump the response to %s", path)
	}

	logger.Logger.Info().Msgf("The deployment result has been dumped to %s", path)

	return nil
}

// name identifies this deployment in a run. On-demand commands fall back into their service names.
func (d *Dumper) name() string {
	if d.deploymentName != "" {
		return d.deploymentName
	}

	return d.serviceName
}

// fileName keeps a deployment name, that is an arbitrary key in a config file, within the run directory.
func fileName(name string) string {
	replacer := strings.NewReplacer("/", "_", `\`, "_", ":", "_")

	if sanitized := strings.Trim(replacer.Replace(name), ". "); sanitized != "" {
		return sanitized + ".json"
	}

	return "result.json"
}

// rawMessage keeps a response as-is if it's a json, otherwise it's embedded as a json string.
func rawMessage(value string) json.RawMessage {
	if json.Valid([]byte(value)) {
		return json.RawMessage(value)
	}

	if bytes, err := json.Marshal(value); err == nil {
		return bytes
	}

	return json.RawMessage("null")
}
