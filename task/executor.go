package task

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"k8s.io/utils/exec"
	"strings"
)

var sh = exec.New()

type Executor struct {
	config *config.ExecutionConfig
	ctx    context.Context
}

func NewExecutor(ctx context.Context, conf *config.ExecutionConfig) *Executor {
	return &Executor{
		config: conf,
		ctx:    ctx,
	}
}

func (p *Executor) Execute(f func() error) error {
	if p.config != nil && len(p.config.PreSteps) > 0 {
		logger.Logger.Info().Msgf("Execute pre-steps... 0/%d", len(p.config.PreSteps))

		if err := runSteps(p.ctx, p.config.PreSteps); err != nil {
			return errors.Wrap(err, "failed to execute pre-steps")
		}
	} else {
		logger.Logger.Info().Msg("No pre-steps are found")
	}

	if err := f(); err != nil {
		return errors.Wrap(err, "failed to execute the content so post-steps won't be executed.")
	}

	if p.config != nil && len(p.config.PostSteps) > 0 {
		logger.Logger.Info().Msgf("Execute post-steps... 0/%d", len(p.config.PostSteps))

		if err := runSteps(p.ctx, p.config.PostSteps); err != nil {
			return errors.Wrap(err, "failed to execute post-steps")
		}
	} else {
		logger.Logger.Info().Msg("No post-steps are found")
	}

	return nil
}

func runSteps(ctx context.Context, steps [][]string) error {
	for idx, args := range steps {
		logger.Logger.Info().Msgf("Start executing steps... %d/%d", idx+1, len(steps))

		cmd := sh.CommandContext(ctx, args[0], args[1:]...)
		cmd.SetStdout(logger.CmdStdout)
		cmd.SetStderr(logger.CmdStderr)

		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "failed to execute %s... %d/%d", strings.Join(args, " "), idx+1, len(steps))
		}

		logger.Logger.Info().Msgf("Finished steps... %d/%d", idx+1, len(steps))
	}

	return nil
}
