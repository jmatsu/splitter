package task

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
)

func DeployToCustomService(ctx context.Context, deploymentName string, serviceName string, def config.CustomServiceDefinition, conf config.CustomServiceConfig, filePath string, builder func(req *service.CustomServiceDeployRequest) error) error {
	if err := conf.Validate(); err != nil {
		return errors.Wrap(err, "the built config is invalid")
	}

	provider := service.NewCustomServiceProvider(ctx, &def, &conf)

	formatter := NewFormatter()
	formatter.TableBuilder = nil

	dumper := NewDumper(serviceName, deploymentName, filePath)

	response, err := provider.Deploy(filePath, builder)

	if err != nil {
		return errors.Wrap(err, "cannot deploy this app")
	}

	if err := formatter.Format(response); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	if err := dumper.Dump(response); err != nil {
		return errors.Wrap(err, "cannot dump the response")
	}

	return nil
}
