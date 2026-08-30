package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmatsu/splitter/internal/config"
)

func Test_LocalProvider_Distribute(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		DestinationExists bool

		Overwrite    bool
		DeleteSource bool
		FileMode     os.FileMode

		expected struct {
			SideEffect sideEffect
		}
	}{
		"copy and overwrite": {
			DestinationExists: true,

			Overwrite:    true,
			DeleteSource: false,
			FileMode:     0644,

			expected: struct {
				SideEffect sideEffect
			}{SideEffect: localCopyAndOverwrite},
		},
		"copy but do not overwrite": {
			DestinationExists: false,

			Overwrite:    true,
			DeleteSource: false,

			expected: struct {
				SideEffect sideEffect
			}{SideEffect: localCopyOnly},
		},
		"copy but can not overwrite": {
			DestinationExists: true,

			Overwrite:    false,
			DeleteSource: false,
			FileMode:     0644,

			expected: struct {
				SideEffect sideEffect
			}{SideEffect: "none"},
		},
		"move and overwrite": {
			DestinationExists: true,

			Overwrite:    true,
			DeleteSource: true,
			FileMode:     0644,

			expected: struct {
				SideEffect sideEffect
			}{SideEffect: localMoveAndOverwrite},
		},
		"move but do not overwrite": {
			DestinationExists: false,

			Overwrite:    true,
			DeleteSource: true,

			expected: struct {
				SideEffect sideEffect
			}{SideEffect: localMoveOnly},
		},
		"move but can not overwrite": {
			DestinationExists: true,

			Overwrite:    false,
			DeleteSource: true,
			FileMode:     0644,

			expected: struct {
				SideEffect sideEffect
			}{SideEffect: "none"},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			source, _ := os.CreateTemp("", "source-*")
			defer func() {
				_ = source.Close()
			}()

			if c.FileMode != 0 {
				if err := source.Chmod(c.FileMode); err != nil {
					t.Errorf("failed to chmod: %v", err)
					return
				}
			}

			stat, err := source.Stat()

			if err != nil {
				t.Errorf("failed to stat: %v", err)
				return
			}

			actual := struct {
				FileMode os.FileMode
			}{
				FileMode: stat.Mode(),
			}

			dest, _ := os.CreateTemp("", "dest-*")

			if c.DestinationExists {
				defer func() {
					_ = dest.Close()
				}()
			} else {
				_ = dest.Close()
				_ = os.Remove(dest.Name())
			}

			provider := NewLocalProvider(context.TODO(), &config.LocalConfig{
				AllowOverwrite:  c.Overwrite,
				DeleteSource:    c.DeleteSource,
				FileMode:        c.FileMode,
				DestinationPath: dest.Name(),
			})

			response, err := provider.Deploy(source.Name())

			if err != nil {
				if !c.Overwrite && strings.Contains(err.Error(), "overwriting is disabled") {
					return // OK
				}

				t.Errorf("failed to distribute: %v", err)
				return
			}

			if response.SideEffect != c.expected.SideEffect {
				t.Errorf("expected to be %s but %s", c.expected.SideEffect, response.SideEffect)
			}

			stat, err = os.Stat(dest.Name()) // Do not use dest.Stat() to get the latest stats.

			if err != nil {
				t.Errorf("failed to stat: %v", err)
				return
			}

			if c.DeleteSource {
				if _, err := os.Stat(source.Name()); err == nil {
					t.Errorf("failed to delete the source file")
					return
				}
			} else {
				if _, err := os.Stat(source.Name()); err != nil {
					t.Errorf("failed to keep the source file")
					return
				}
			}

			if c.FileMode == 0 {
				if stat.Mode() != actual.FileMode {
					t.Errorf("failed to keep the mode of the source file")
					return
				}
			} else if c.FileMode != stat.Mode() {
				t.Errorf("failed to change the mode of the source file properly")
			}
		})
	}
}

func Test_LocalProvider_Deploy_result(t *testing.T) {
	t.Parallel()

	source := newSourceFile(t, "app.apk", "app content")
	destination := filepath.Join(t.TempDir(), "dist", "app.apk")

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		t.Fatalf("failed to prepare the destination: %v", err)
	}

	provider := NewLocalProvider(context.TODO(), &config.LocalConfig{
		DestinationPath: destination,
	})

	result, err := provider.Deploy(source)

	if err != nil {
		t.Fatalf("failed to deploy: %v", err)
	}

	if result.SourceFilePath != source {
		t.Errorf("source path is expected to be %s but %s", source, result.SourceFilePath)
	}

	if result.DestinationFilePath != destination {
		t.Errorf("destination path is expected to be %s but %s", destination, result.DestinationFilePath)
	}

	if result.SideEffect != localCopyOnly {
		t.Errorf("side effect is expected to be %s but %s", localCopyOnly, result.SideEffect)
	}

	var raw LocalMoveResponse

	if err := json.Unmarshal([]byte(result.RawJsonResponse()), &raw); err != nil {
		t.Errorf("the raw response is expected to be a json but %s", result.RawJsonResponse())
	} else if raw.SideEffect != localCopyOnly {
		t.Errorf("the raw response is expected to hold the side effect but %s", raw.SideEffect)
	}

	if _, ok := result.ValueResponse().(LocalDeployResult); !ok {
		t.Errorf("the value response is expected to be LocalDeployResult but %T", result.ValueResponse())
	}

	if bytes, err := os.ReadFile(destination); err != nil {
		t.Errorf("failed to read the destination: %v", err)
	} else if string(bytes) != "app content" {
		t.Errorf("the destination is expected to hold the source content but %s", string(bytes))
	}
}

func Test_LocalProvider_Deploy_failures(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		sourceExists      bool
		destinationIsDir  bool
		allowOverwrite    bool
		destinationExists bool
	}{
		"missing source": {
			sourceExists: false,
		},
		"a directory as a destination": {
			sourceExists:      true,
			destinationIsDir:  true,
			destinationExists: true,
			allowOverwrite:    true,
		},
		"overwriting is disabled": {
			sourceExists:      true,
			destinationExists: true,
			allowOverwrite:    false,
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			source := filepath.Join(dir, "app.apk")

			if c.sourceExists {
				if err := os.WriteFile(source, []byte("app content"), 0644); err != nil {
					t.Fatalf("failed to prepare the source: %v", err)
				}
			}

			destination := filepath.Join(dir, "dest")

			if c.destinationExists {
				if c.destinationIsDir {
					if err := os.MkdirAll(destination, 0755); err != nil {
						t.Fatalf("failed to prepare the destination: %v", err)
					}
				} else if err := os.WriteFile(destination, []byte("existing"), 0644); err != nil {
					t.Fatalf("failed to prepare the destination: %v", err)
				}
			}

			provider := NewLocalProvider(context.TODO(), &config.LocalConfig{
				DestinationPath: destination,
				AllowOverwrite:  c.allowOverwrite,
			})

			if _, err := provider.Deploy(source); err == nil {
				t.Errorf("%s case is expected to be failure but not", name)
			}
		})
	}
}
