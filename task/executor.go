package task

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"k8s.io/utils/exec"
	"strings"
)

// Executor is the environment to run commands.
type Executor struct {
	config *config.ExecutionConfig
	ctx    context.Context
	sh     exec.Interface
	stdout io.Writer
	stderr io.Writer
}

type commandWriter struct {
	l     *zerolog.Logger
	level zerolog.Level
}

func NewExecutor(sh exec.Interface, ctx context.Context, conf *config.ExecutionConfig) *Executor {
	if sh == nil {
		sh = exec.New()
	}

	return &Executor{
		config: conf,
		ctx:    ctx,
		sh:     sh,
		stdout: newCommandWriter(zerolog.InfoLevel),
		stderr: newCommandWriter(zerolog.WarnLevel),
	}
}

func (e *Executor) Execute(f func() error) error {
	if e.config != nil && len(e.config.PreSteps) > 0 {
		logger.Logger.Info().Msgf("Execute pre-steps... 0/%d", len(e.config.PreSteps))

		if err := e.runSteps(e.config.PreSteps); err != nil {
			return errors.Wrap(err, "failed to execute pre-steps")
		}
	} else {
		logger.Logger.Info().Msg("No pre-steps are found")
	}

	if err := f(); err != nil {
		return errors.Wrap(err, "failed to execute the content so post-steps won't be executed.")
	}

	if e.config != nil && len(e.config.PostSteps) > 0 {
		logger.Logger.Info().Msgf("Execute post-steps... 0/%d", len(e.config.PostSteps))

		if err := e.runSteps(e.config.PostSteps); err != nil {
			return errors.Wrap(err, "failed to execute post-steps")
		}
	} else {
		logger.Logger.Info().Msg("No post-steps are found")
	}

	return nil
}

func (e *Executor) newCmd(args ...string) exec.Cmd {
	cmd := e.sh.CommandContext(e.ctx, args[0], args[1:]...)
	cmd.SetStdout(e.stdout)
	cmd.SetStderr(e.stderr)
	return cmd
}

func (e *Executor) runSteps(steps [][]string) error {
	for idx, args := range steps {
		logger.Logger.Info().Msgf("Start executing steps... %d/%d", idx+1, len(steps))

		cmd := e.newCmd(args...)

		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "failed to execute %s... %d/%d", strings.Join(args, " "), idx+1, len(steps))
		}

		logger.Logger.Info().Msgf("Finished steps... %d/%d", idx+1, len(steps))
	}

	return nil
}

func newCommandWriter(level zerolog.Level) commandWriter {
	l := logger.Logger.Level(level)

	return commandWriter{
		l:     &l,
		level: level,
	}
}

func (l commandWriter) Write(p []byte) (n int, err error) {
	// borrowed from zerolog's code. this code chunk's copyrights belong to zerolog authors.
	// ref: https://github.com/rs/zerolog/blob/3543e9d94bc5ed088dd2d9ad1d19c7ccd0fa65f5/log.go#L435
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		p = p[0 : n-1]
	}
	l.l.WithLevel(l.level).CallerSkipFrame(1).Msg(string(p))
	return
}
