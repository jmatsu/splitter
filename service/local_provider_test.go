package service

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"os"
	"strings"
	"testing"
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
			defer source.Close()

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
				defer dest.Close()
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
