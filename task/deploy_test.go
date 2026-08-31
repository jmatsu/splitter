package task

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
)

// withFormatStyle pins the global format style so that the table builders are actually exercised.
func withFormatStyle(t *testing.T, style config.FormatStyle) {
	t.Helper()

	original := config.CurrentConfig().FormatStyle()

	config.SetGlobalFormatStyle(style)

	t.Cleanup(func() {
		config.SetGlobalFormatStyle(original)
	})
}

// withDistDir pins the dump destination so that tests never write into the working directory.
func withDistDir(t *testing.T) string {
	t.Helper()

	originalDir := config.CurrentConfig().DistDir()
	originalRunName := config.CurrentConfig().RunName()

	dir := t.TempDir()

	config.SetGlobalDistDir(dir)
	config.SetGlobalRunName("run1")

	t.Cleanup(func() {
		config.SetGlobalDistDir(originalDir)
		config.SetGlobalRunName(originalRunName)
	})

	return filepath.Join(dir, "run1")
}

func newSourceFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create %s: %v", path, err)
	}

	return path
}

func Test_DeployToLocal(t *testing.T) {
	for _, style := range []config.FormatStyle{config.PrettyFormat, config.RawFormat, config.MarkdownFormat} {
		style := style

		t.Run(style, func(t *testing.T) {
			withFormatStyle(t, style)

			runDir := withDistDir(t)

			source := newSourceFile(t, "app.apk", "app content")
			destination := filepath.Join(t.TempDir(), "app.apk")

			conf := config.LocalConfig{DestinationPath: destination}

			if err := DeployToLocal(context.TODO(), "local1", conf, source); err != nil {
				t.Fatalf("failed to deploy: %v", err)
			}

			if _, err := os.Stat(filepath.Join(runDir, "local1.json")); err != nil {
				t.Errorf("the result is expected to be dumped but not: %v", err)
			}

			if bytes, err := os.ReadFile(destination); err != nil {
				t.Errorf("the destination is expected to exist but not: %v", err)
			} else if string(bytes) != "app content" {
				t.Errorf("the destination is expected to hold the source content but %s", string(bytes))
			}
		})
	}
}

func Test_DeployToLocal_failures(t *testing.T) {
	withFormatStyle(t, config.PrettyFormat)

	t.Run("invalid config", func(t *testing.T) {
		source := newSourceFile(t, "app.apk", "app content")

		if err := DeployToLocal(context.TODO(), "local1", config.LocalConfig{}, source); err == nil {
			t.Errorf("an invalid config is expected to be rejected but not")
		}
	})

	t.Run("missing source", func(t *testing.T) {
		conf := config.LocalConfig{DestinationPath: filepath.Join(t.TempDir(), "app.apk")}

		if err := DeployToLocal(context.TODO(), "local1", conf, filepath.Join(t.TempDir(), "not-found.apk")); err == nil {
			t.Errorf("a missing source is expected to be rejected but not")
		}
	})
}

func Test_DeployToCustomService(t *testing.T) {
	withFormatStyle(t, config.RawFormat) // a custom service has no table builder

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"value"}`))
	}))

	t.Cleanup(server.Close)

	definition := config.CustomServiceDefinition{
		Endpoint:         server.URL + "/path/to/upload",
		SourceFileFormat: config.FormParamsAssignFormatPrefix + "file",
		AuthDefinition: config.CustomAuthDefinition{
			StyleFormat: config.HeadersAssignFormatPrefix + "Authorization",
			ValueFormat: "Bearer %s",
		},
	}

	source := newSourceFile(t, "app.apk", "app content")

	t.Run("success", func(t *testing.T) {
		runDir := withDistDir(t)

		conf := config.CustomServiceConfig{AuthToken: "token1"}

		if err := DeployToCustomService(context.TODO(), "", "custom1", definition, conf, source, func(req *service.CustomServiceDeployRequest) error {
			return nil
		}); err != nil {
			t.Errorf("failed to deploy: %v", err)
		}

		// An on-demand deployment has no deployment name so the service name is used instead.
		if _, err := os.Stat(filepath.Join(runDir, "custom1.json")); err != nil {
			t.Errorf("the result is expected to be dumped but not: %v", err)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		if err := DeployToCustomService(context.TODO(), "", "custom1", definition, config.CustomServiceConfig{}, source, func(req *service.CustomServiceDeployRequest) error {
			return nil
		}); err == nil {
			t.Errorf("an invalid config is expected to be rejected but not")
		}
	})
}

// The remaining services talk to their own hosts so only the config validation is verifiable here.
func Test_Deploy_invalidConfigs(t *testing.T) {
	withFormatStyle(t, config.PrettyFormat)

	source := newSourceFile(t, "app.apk", "app content")

	t.Run("deploygate", func(t *testing.T) {
		if err := DeployToDeployGate(context.TODO(), "", config.DeployGateConfig{}, source, func(req *service.DeployGateDeployRequest) error {
			return nil
		}); err == nil {
			t.Errorf("an invalid config is expected to be rejected but not")
		}
	})

	t.Run("firebase app distribution", func(t *testing.T) {
		if err := DeployToFirebaseAppDistribution(context.TODO(), "", config.FirebaseAppDistributionConfig{}, source, func(req *service.FirebaseAppDistributionDeployRequest) error {
			return nil
		}); err == nil {
			t.Errorf("an invalid config is expected to be rejected but not")
		}
	})

	t.Run("test flight", func(t *testing.T) {
		if err := DeployToTestFlight(context.TODO(), "", config.TestFlightConfig{}, source, func(req *service.TestFlightDeployRequest) error {
			return nil
		}); err == nil {
			t.Errorf("an invalid config is expected to be rejected but not")
		}
	})
}
