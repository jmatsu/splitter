package task

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/util"
	"github.com/pkg/errors"
	"k8s.io/utils/exec"
	"strings"
)

// StepExecutor is the environment to run commands.
type StepExecutor struct {
	config      *config.ExecutionConfig
	commandLine util.CommandLine
}

func NewExecutor(ctx context.Context, sh exec.Interface, conf *config.ExecutionConfig) *StepExecutor {
	return &StepExecutor{
		config:      conf,
		commandLine: util.NewCommandLine(ctx, sh),
	}
}

func (e *StepExecutor) Execute(f func() error) error {
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

func (e *StepExecutor) runSteps(steps [][]string) error {
	for idx, args := range steps {
		logger.Logger.Info().Msgf("Start executing steps... %d/%d", idx+1, len(steps))

		var err error

		if len(args) >= 2 {
			_, _, err = e.commandLine.Exec(args[0], args[1:]...)
		} else {
			_, _, err = e.commandLine.Exec(args[0])
		}

		if err != nil {
			return errors.Wrapf(err, "failed to execute %s... %d/%d", strings.Join(args, " "), idx+1, len(steps))
		}

		logger.Logger.Info().Msgf("Finished steps... %d/%d", idx+1, len(steps))
	}

	return nil
}
