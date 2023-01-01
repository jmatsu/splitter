package lifecycle

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"k8s.io/utils/exec"
	"strings"
)

var sh = exec.New()

type Provider struct {
	config *config.LifecycleConfig
	ctx    context.Context
}

func NewProvider(ctx context.Context, conf *config.LifecycleConfig) *Provider {
	return &Provider{
		config: conf,
		ctx:    ctx,
	}
}

func (p *Provider) Execute(f func() error) error {
	if err := p.RunPreSteps(); err != nil {
		return errors.Wrap(err, "failed to execute pre-steps")
	}

	if err := f(); err != nil {
		return errors.Wrap(err, "failed to execute the content so post-steps won't be executed.")
	}

	if err := p.RunPostSteps(); err != nil {
		return errors.Wrap(err, "failed to execute post-steps")
	}

	return nil
}

func (p *Provider) RunPreSteps() error {
	if p.config == nil && len(p.config.PreSteps) == 0 {
		return nil
	}

	logger.Logger.Info().Msgf("Execute pre-steps... 0/%d", len(p.config.PreSteps))

	return runSteps(p.ctx, p.config.PreSteps)
}

func (p *Provider) RunPostSteps() error {
	if p.config == nil && len(p.config.PostSteps) == 0 {
		return nil
	}

	logger.Logger.Info().Msgf("Execute post-steps... 0/%d", len(p.config.PostSteps))

	return runSteps(p.ctx, p.config.PostSteps)
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
