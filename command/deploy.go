package command

import (
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
	"github.com/urfave/cli/v2"
)

// Deploy command distributes your app to pre-defined services in your config file.
func Deploy(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Manage your apps' deployments with following the configuration.",
		Description: "You can deploy your apps to supported services based on pre-defined service configuration.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "name",
				Aliases: []string{
					"n",
				},
				Usage:    "deployment name in your configuration file.",
				Required: true,
				EnvVars:  []string{config.ToEnvName("DEPLOYMENT_NAME")},
			},
			&cli.PathFlag{
				Name: "source-path",
				Aliases: []string{
					"f",
				},
				Usage:    "A path to an app file.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "release-note",
				Usage:    "An release note of this revision. Some of services may not support this option.",
				Required: false,
				EnvVars:  []string{config.ToEnvName("DEPLOYMENT_RELEASE_NOTE")},
			},
		},
		Action: func(context *cli.Context) error {
			name := context.String("name")

			logger.Logger.Info().Msgf("Loading %s config...", name)

			deployment, definition, err := config.CurrentConfig().Deployment(name)

			if err != nil {
				return err
			}

			executor := task.NewExecutor(context.Context, nil, &deployment.Lifecycle)

			return executor.Execute(func() error {
				sourceFilePath := context.String("source-path")

				switch deployment.ServiceName {
				case config.DeploygateService:
					dg := deployment.ServiceConfig.(config.DeployGateConfig)

					return task.DeployToDeployGate(context.Context, dg, sourceFilePath, func(req *service.DeployGateDeployRequest) error {
						if v := context.String("release-note"); context.IsSet("release-note") {
							req.SetMessage(v)
							req.SetDistributionReleaseNote(v)
						}

						return nil
					})
				case config.LocalService:
					lo := deployment.ServiceConfig.(config.LocalConfig)

					return task.DeployToLocal(context.Context, lo, sourceFilePath)
				case config.FirebaseAppDistributionService:
					fad := deployment.ServiceConfig.(config.FirebaseAppDistributionConfig)

					return task.DeployToFirebaseAppDistribution(context.Context, fad, sourceFilePath, func(req *service.FirebaseAppDistributionDeployRequest) error {
						if v := context.String("release-note"); context.IsSet("release-note") {
							req.SetReleaseNote(v)
						}

						return nil
					})
				default:
					custom := deployment.ServiceConfig.(config.CustomServiceConfig)

					return task.DeployToCustomService(context.Context, definition, custom, sourceFilePath, func(req *service.CustomServiceDeployRequest) error {
						return nil
					})
				}
			})
		},
	}
}
