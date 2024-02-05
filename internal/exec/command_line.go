package exec

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"k8s.io/utils/exec"
)

type CommandLine interface {
	Exec(command string, args ...string) ([]byte, []byte, error)
}

type commandExecutor struct {
	ctx       context.Context
	execution exec.Interface
	stdout    io.Writer
	stderr    io.Writer
}

func NewCommandLine(ctx context.Context, execution exec.Interface) CommandLine {
	if execution == nil {
		execution = exec.New()
	}

	return &commandExecutor{
		ctx:       ctx,
		execution: execution,
		stdout:    withLogger(&logger.Logger, zerolog.InfoLevel),
		stderr:    withLogger(&logger.Logger, zerolog.WarnLevel),
	}
}

func (s *commandExecutor) Exec(command string, args ...string) ([]byte, []byte, error) {
	cmd := s.execution.CommandContext(s.ctx, command, args...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	mout := mirroring{
		writers: []io.Writer{
			&stdout,
			s.stdout,
		},
	}
	merr := mirroring{
		writers: []io.Writer{
			&stderr,
			s.stderr,
		},
	}

	cmd.SetStdout(mout)
	cmd.SetStderr(merr)

	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), stderr.Bytes(), errors.Wrapf(err, fmt.Sprintf("%s failed to run", command))
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

type mirroring struct {
	writers []io.Writer
}

func (m mirroring) Write(p []byte) (n int, err error) {
	for _, w := range m.writers {
		n, err = w.Write(p)

		if err != nil {
			return
		}
	}

	return
}

type leveledWriter struct {
	l     *zerolog.Logger
	level zerolog.Level
}

func withLogger(logger *zerolog.Logger, level zerolog.Level) leveledWriter {
	l := logger.Level(level)

	return leveledWriter{
		l:     &l,
		level: level,
	}
}

func (l leveledWriter) Write(p []byte) (n int, err error) {
	// borrowed from zerolog's code. this code chunk's copyrights belong to zerolog authors.
	// ref: https://github.com/rs/zerolog/blob/3543e9d94bc5ed088dd2d9ad1d19c7ccd0fa65f5/log.go#L435
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		p = p[0 : n-1]
	}
	l.l.WithLevel(l.level).CallerSkipFrame(1).Msg(string(p))
	return
}
