// Package tests holds end-to-end tests that drive the built splitter binary.
//
// Only the services that need no credentials are covered here. The others are exercised by
// .github/workflows/integration-test.yml with the repository secrets.
package tests

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var splitterPath string

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		os.Exit(m.Run())
	}

	dir, err := os.MkdirTemp("", "splitter-e2e-*")

	if err != nil {
		fmt.Printf("failed to create a temp dir: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	splitterPath = filepath.Join(dir, "splitter")

	build := exec.Command("go", "build", "-o", splitterPath, "..")

	if out, err := build.CombinedOutput(); err != nil {
		fmt.Printf("failed to build splitter: %v\n%s\n", err, string(out))
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type execution struct {
	stdout   string
	stderr   string
	exitCode int
}

func (e execution) combined() string {
	return e.stdout + e.stderr
}

// run executes the built binary. It never fails the test so that error cases are assertable too.
func run(t *testing.T, env []string, args ...string) execution {
	t.Helper()

	if testing.Short() {
		t.Skip("e2e tests are skipped in the short mode")
	}

	cmd := exec.Command(splitterPath, args...)
	cmd.Env = append(os.Environ(), append([]string{"SPLITTER_LOG_LEVEL=info"}, env...)...)
	cmd.Dir = t.TempDir() // keep the runs away from the config file of this repository

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := execution{stdout: stdout.String(), stderr: stderr.String()}

	if err != nil {
		var exitErr *exec.ExitError

		if !errors.As(err, &exitErr) {
			t.Fatalf("failed to run splitter: %v", err)
		}

		result.exitCode = exitErr.ExitCode()
	}

	return result
}

func writeFile(t *testing.T, path string, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}

	return path
}

func assertFileContent(t *testing.T, path string, expected string) {
	t.Helper()

	bytes, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	if string(bytes) != expected {
		t.Errorf("%s is expected to hold %q but %q", path, expected, string(bytes))
	}
}

func Test_E2E_help(t *testing.T) {
	result := run(t, nil, "--help")

	if result.exitCode != 0 {
		t.Fatalf("--help is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}

	for _, name := range []string{"init", "local", "deploygate", "firebase-app-distribution", "deploy", "add-deployment", "service", "test-flight"} {
		if !strings.Contains(result.stdout, name) {
			t.Errorf("%s command is expected to be listed but not:\n%s", name, result.stdout)
		}
	}
}

func Test_E2E_local(t *testing.T) {
	cases := map[string]struct {
		destinationExists bool
		args              []string

		expectedExitCode int
		expectedContent  string
		sourceKept       bool
		expectedFileMode os.FileMode
	}{
		"copy": {
			args:            []string{},
			expectedContent: "app content",
			sourceKept:      true,
		},
		"copy and overwrite": {
			destinationExists: true,
			args:              []string{"--overwrite"},
			expectedContent:   "app content",
			sourceKept:        true,
		},
		"move": {
			args:            []string{"--delete-source"},
			expectedContent: "app content",
		},
		"with a file mode": {
			args:             []string{"--file-mode", "384"}, // 0600
			expectedContent:  "app content",
			sourceKept:       true,
			expectedFileMode: 0600,
		},
		"overwriting is disabled": {
			destinationExists: true,
			args:              []string{},
			expectedExitCode:  1,
			expectedContent:   "existing content",
			sourceKept:        true,
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()

			source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")
			destination := filepath.Join(dir, "dist", "app.apk")

			if c.destinationExists {
				writeFile(t, destination, "existing content")
			} else if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
				t.Fatalf("failed to create the destination dir: %v", err)
			}

			args := append([]string{"local", "--source-path", source, "--destination-path", destination}, c.args...)

			result := run(t, nil, args...)

			if result.exitCode != c.expectedExitCode {
				t.Fatalf("exit code is expected to be %d but %d: %s", c.expectedExitCode, result.exitCode, result.combined())
			}

			assertFileContent(t, destination, c.expectedContent)

			if _, err := os.Stat(source); (err == nil) != c.sourceKept {
				t.Errorf("the source is expected to be kept %t but %t", c.sourceKept, err == nil)
			}

			if c.expectedFileMode != 0 {
				if stat, err := os.Stat(destination); err != nil {
					t.Errorf("failed to stat the destination: %v", err)
				} else if stat.Mode() != c.expectedFileMode {
					t.Errorf("file mode is expected to be %s but %s", c.expectedFileMode, stat.Mode())
				}
			}
		})
	}
}

func Test_E2E_init(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "splitter.yml")

	if result := run(t, nil, "init", "--path", path); result.exitCode != 0 {
		t.Fatalf("init is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s is expected to be generated but not: %v", path, err)
	}

	// The second run must not clobber the existing file silently.
	if result := run(t, nil, "init", "--path", path); result.exitCode == 0 {
		t.Errorf("init is expected to reject the existing file but not: %s", result.combined())
	}

	if result := run(t, nil, "init", "--path", path, "--overwrite"); result.exitCode != 0 {
		t.Errorf("init --overwrite is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}
}

func Test_E2E_addDeployment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "splitter.yml")

	writeFile(t, path, "deployments: {}\n")

	result := run(t, nil, "--config", path, "add-deployment", "--path", path, "--name", "new1", "--service", "local")

	if result.exitCode != 0 {
		t.Fatalf("add-deployment is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}

	bytes, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	for _, expected := range []string{"new1", "service: local", "destination-path"} {
		if !strings.Contains(string(bytes), expected) {
			t.Errorf("%s is expected to be in the config but not:\n%s", expected, string(bytes))
		}
	}

	// A name and a service are mandatory.
	if result := run(t, nil, "--config", path, "add-deployment", "--path", path, "--service", "local"); result.exitCode == 0 {
		t.Errorf("add-deployment without a name is expected to fail but not")
	}

	if result := run(t, nil, "--config", path, "add-deployment", "--path", path, "--name", "new2"); result.exitCode == 0 {
		t.Errorf("add-deployment without a service is expected to fail but not")
	}

	// The same name must not be added twice.
	if result := run(t, nil, "--config", path, "add-deployment", "--path", path, "--name", "new1", "--service", "local"); result.exitCode == 0 {
		t.Errorf("add-deployment is expected to reject the duplicated name but not")
	}
}

func Test_E2E_deployToLocal(t *testing.T) {
	cases := map[string]struct {
		format string
	}{
		"pretty":   {format: "pretty"},
		"raw":      {format: "raw"},
		"markdown": {format: "markdown"},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()

			source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")
			destination := filepath.Join(dir, "dist", "app.apk")

			if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
				t.Fatalf("failed to create the destination dir: %v", err)
			}

			config := writeFile(t, filepath.Join(dir, "splitter.yml"), fmt.Sprintf(`
deployments:
  local1:
    service: local
    destination-path: %s
    allow-overwrite: true
`, destination))

			result := run(t, nil, "--config", config, "--format", c.format, "deploy", "-n", "local1", "-f", source)

			if result.exitCode != 0 {
				t.Fatalf("deploy is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
			}

			assertFileContent(t, destination, "app content")

			if !strings.Contains(result.combined(), destination) {
				t.Errorf("the destination is expected to be reported but not:\n%s", result.combined())
			}
		})
	}
}

func Test_E2E_deployWithLifecycleSteps(t *testing.T) {
	dir := t.TempDir()

	source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")
	destination := filepath.Join(dir, "dist", "app.apk")
	copied := filepath.Join(dir, "dist", "app.copied.apk")

	config := writeFile(t, filepath.Join(dir, "splitter.yml"), fmt.Sprintf(`
deployments:
  local1:
    service: local
    destination-path: %s
    allow-overwrite: true
    pre-steps:
      - ["mkdir", "-p", "%s"]
    post-steps:
      - ["cp", "-f", "%s", "%s"]
`, destination, filepath.Dir(destination), destination, copied))

	result := run(t, nil, "--config", config, "deploy", "-n", "local1", "-f", source)

	if result.exitCode != 0 {
		t.Fatalf("deploy is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}

	assertFileContent(t, destination, "app content")
	assertFileContent(t, copied, "app content")
}

func Test_E2E_deployWithFailingPreStep(t *testing.T) {
	dir := t.TempDir()

	source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")
	destination := filepath.Join(dir, "dist", "app.apk")

	config := writeFile(t, filepath.Join(dir, "splitter.yml"), fmt.Sprintf(`
deployments:
  local1:
    service: local
    destination-path: %s
    pre-steps:
      - ["false"]
`, destination))

	result := run(t, nil, "--config", config, "deploy", "-n", "local1", "-f", source)

	if result.exitCode == 0 {
		t.Fatalf("a failing pre-step is expected to halt the deployment but not: %s", result.combined())
	}

	if _, err := os.Stat(destination); err == nil {
		t.Errorf("the deployment is expected to be halted but the destination exists")
	}
}

func Test_E2E_deployWithEmbeddedVariables(t *testing.T) {
	dir := t.TempDir()

	source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")
	destination := filepath.Join(dir, "dist", "app.apk")

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		t.Fatalf("failed to create the destination dir: %v", err)
	}

	config := writeFile(t, filepath.Join(dir, "splitter.yml"), `
deployments:
  local1:
    service: local
    destination-path: "format:${E2E_DESTINATION_PATH}"
`)

	result := run(t, []string{"E2E_DESTINATION_PATH=" + destination}, "--config", config, "deploy", "-n", "local1", "-f", source)

	if result.exitCode != 0 {
		t.Fatalf("deploy is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}

	assertFileContent(t, destination, "app content")
}

func Test_E2E_deployToCustomService(t *testing.T) {
	var (
		authorization string
		uploaded      string
		defaultHeader string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		defaultHeader = r.Header.Get("X-Default")

		if err := r.ParseMultipartForm(32 << 20); err == nil && r.MultipartForm != nil {
			if headers := r.MultipartForm.File["file"]; len(headers) > 0 {
				if f, err := headers[0].Open(); err == nil {
					defer func() {
						_ = f.Close()
					}()

					bytes := make([]byte, headers[0].Size)
					_, _ = f.Read(bytes)
					uploaded = string(bytes)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}))

	t.Cleanup(server.Close)

	dir := t.TempDir()

	source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")

	config := writeFile(t, filepath.Join(dir, "splitter.yml"), fmt.Sprintf(`
services:
  custom1:
    endpoint: %s/path/to/upload
    source-file-format: form_params.file
    auth:
      style-format: headers.Authorization
      value-format: "Bearer %%s"
    default:
      headers:
        X-Default: default-header
deployments:
  custom-deployment:
    service: custom1
    auth-token: "format:${E2E_AUTH_TOKEN}"
`, server.URL))

	result := run(t, []string{"E2E_AUTH_TOKEN=token1"}, "--config", config, "--format", "raw", "deploy", "-n", "custom-deployment", "-f", source)

	if result.exitCode != 0 {
		t.Fatalf("deploy is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}

	if authorization != "Bearer token1" {
		t.Errorf("authorization is expected to be Bearer token1 but %s", authorization)
	}

	if defaultHeader != "default-header" {
		t.Errorf("the default header is expected to be sent but %s", defaultHeader)
	}

	if uploaded != "app content" {
		t.Errorf("the source file is expected to be uploaded but %q", uploaded)
	}

	if !strings.Contains(result.stdout, `{"result":"ok"}`) {
		t.Errorf("the raw response is expected to be printed but not:\n%s", result.stdout)
	}
}

func Test_E2E_serviceCommand(t *testing.T) {
	var (
		authorization string
		query         string
		formParam     string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		query = r.URL.RawQuery

		if err := r.ParseMultipartForm(32 << 20); err == nil && r.MultipartForm != nil {
			if values := r.MultipartForm.Value["extra"]; len(values) > 0 {
				formParam = values[0]
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}))

	t.Cleanup(server.Close)

	dir := t.TempDir()

	source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")

	config := writeFile(t, filepath.Join(dir, "splitter.yml"), fmt.Sprintf(`
services:
  custom1:
    endpoint: %s/path/to/upload
    source-file-format: form_params.file
    auth:
      style-format: headers.Authorization
      value-format: "Bearer %%s"
`, server.URL))

	result := run(t, nil,
		"--config", config, "--format", "raw",
		"service",
		"--name", "custom1",
		"--auth-token", "token1",
		"--source-path", source,
		"--query-param", "key1=value1",
		"--form-param", "extra=extra-value",
	)

	if result.exitCode != 0 {
		t.Fatalf("service is expected to exit with 0 but %d: %s", result.exitCode, result.combined())
	}

	if authorization != "Bearer token1" {
		t.Errorf("authorization is expected to be Bearer token1 but %s", authorization)
	}

	if query != "key1=value1" {
		t.Errorf("query is expected to be key1=value1 but %s", query)
	}

	if formParam != "extra-value" {
		t.Errorf("the form param is expected to be extra-value but %s", formParam)
	}
}

func Test_E2E_failures(t *testing.T) {
	dir := t.TempDir()

	source := writeFile(t, filepath.Join(dir, "app.apk"), "app content")

	config := writeFile(t, filepath.Join(dir, "splitter.yml"), `
deployments:
  local1:
    service: local
    destination-path: ./dist/app.apk
`)

	cases := map[string]struct {
		args []string
	}{
		"unknown deployment": {
			args: []string{"--config", config, "deploy", "-n", "unknown", "-f", source},
		},
		"missing config file": {
			args: []string{"--config", filepath.Join(dir, "not-found.yml"), "deploy", "-n", "local1", "-f", source},
		},
		"unknown format style": {
			args: []string{"--config", config, "--format", "obababa", "deploy", "-n", "local1", "-f", source},
		},
		"too long network timeout": {
			args: []string{"--config", config, "--network-timeout", "31m", "deploy", "-n", "local1", "-f", source},
		},
		"malformed wait timeout": {
			args: []string{"--config", config, "--wait-timeout", "5 minutes", "deploy", "-n", "local1", "-f", source},
		},
		"unknown service in the config": {
			args: []string{
				"--config", writeFile(t, filepath.Join(dir, "unknown-service.yml"), "deployments:\n  d1:\n    service: obababa\n"),
				"deploy", "-n", "d1", "-f", source,
			},
		},
		"missing required values": {
			args: []string{
				"--config", writeFile(t, filepath.Join(dir, "lacked.yml"), "deployments:\n  d1:\n    service: deploygate\n"),
				"deploy", "-n", "d1", "-f", source,
			},
		},
		"unknown command": {
			args: []string{"obababa"},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			if result := run(t, []string{"DEPLOYGATE_API_TOKEN=", "DEPLOYGATE_APP_OWNER_NAME="}, c.args...); result.exitCode == 0 {
				t.Errorf("%s case is expected to exit with non-zero but 0:\n%s", name, result.combined())
			}
		})
	}
}
