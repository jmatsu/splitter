package task

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmatsu/splitter/service"
)

func Test_Dumper_Dump(t *testing.T) {
	cases := map[string]struct {
		deploymentName   string
		expectedFileName string
	}{
		"with a deployment name": {
			deploymentName:   "pull-request",
			expectedFileName: "pull-request.json",
		},
		"without a deployment name": {
			expectedFileName: "local.json",
		},
		"with a deployment name that contains separators": {
			deploymentName:   "pulls/1234",
			expectedFileName: "pulls_1234.json",
		},
		"with a deployment name that traverses the parent": {
			deploymentName:   "..",
			expectedFileName: "result.json",
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			runDir := withDistDir(t)

			if err := NewDumper("local", c.deploymentName, "app.apk").Dump(&testDistributionResult{}); err != nil {
				t.Fatalf("failed to dump: %v", err)
			}

			if _, err := os.Stat(filepath.Join(runDir, c.expectedFileName)); err != nil {
				t.Errorf("%s is expected to be dumped but not: %v", c.expectedFileName, err)
			}
		})
	}
}

func Test_Dumper_Dump_content(t *testing.T) {
	runDir := withDistDir(t)

	if err := NewDumper("local", "local1", "app.apk").Dump(&testDistributionResult{}); err != nil {
		t.Fatalf("failed to dump: %v", err)
	}

	bytes, err := os.ReadFile(filepath.Join(runDir, "local1.json"))

	if err != nil {
		t.Fatalf("failed to read the dumped file: %v", err)
	}

	var dumped map[string]any

	if err := json.Unmarshal(bytes, &dumped); err != nil {
		t.Fatalf("the dumped file is expected to be a json but %s: %v", string(bytes), err)
	}

	for _, key := range []string{"service", "deployment", "run_name", "deployed_at", "source_file_path", "app", "release", "values", "raw"} {
		if _, found := dumped[key]; !found {
			t.Errorf("%s is expected to be dumped but not: %v", key, dumped)
		}
	}

	if dumped["run_name"] != "run1" {
		t.Errorf("the run name is expected to be run1 but %v", dumped["run_name"])
	}

	// testDistributionResult returns a non-json response so it must be embedded as a json string.
	if dumped["raw"] != "ok" {
		t.Errorf("a non-json response is expected to be embedded as a string but %v", dumped["raw"])
	}
}

func Test_rawMessage(t *testing.T) {
	cases := map[string]struct {
		value    string
		expected string
	}{
		"json object":    {value: `{"key":"value"}`, expected: `{"key":"value"}`},
		"json array":     {value: `[1,2]`, expected: `[1,2]`},
		"malformed json": {value: `{"key":`, expected: `"{\"key\":"`},
		"empty":          {value: "", expected: `""`},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			if v := string(rawMessage(c.value)); v != c.expected {
				t.Errorf("%s is expected to be %s but %s", c.value, c.expected, v)
			}
		})
	}
}

func Test_Dumper_Dump_normalizedResponse(t *testing.T) {
	runDir := withDistDir(t)

	result := service.LocalDeployResult{
		LocalMoveResponse: service.LocalMoveResponse{
			SourceFilePath:      "app.apk",
			DestinationFilePath: "dist/app.apk",
			SideEffect:          "copied",
		},
		RawJson: `{"key":"value"}`,
	}

	if err := NewDumper("local", "local1", "app.apk").Dump(&result); err != nil {
		t.Fatalf("failed to dump: %v", err)
	}

	bytes, err := os.ReadFile(filepath.Join(runDir, "local1.json"))

	if err != nil {
		t.Fatalf("failed to read the dumped file: %v", err)
	}

	var dumped struct {
		App     service.NormalizedApp     `json:"app"`
		Release service.NormalizedRelease `json:"release"`
	}

	if err := json.Unmarshal(bytes, &dumped); err != nil {
		t.Fatalf("the dumped file is expected to be a json but %s: %v", string(bytes), err)
	}

	if dumped.Release.DestinationPath == nil || *dumped.Release.DestinationPath != "dist/app.apk" {
		t.Errorf("the destination path is expected to be dist/app.apk but %v", dumped.Release.DestinationPath)
	}

	// A service that doesn't expose a value must leave the field null instead of an empty string.
	if dumped.App.Name != nil {
		t.Errorf("the app name is expected to be null but %v", *dumped.App.Name)
	}
}
