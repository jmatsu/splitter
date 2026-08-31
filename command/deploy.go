package command

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
	"github.com/urfave/cli/v3"
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
				Sources:  cli.EnvVars(config.ToEnvName("DEPLOYMENT_NAME")),
			},
			&cli.StringFlag{
				Name: "source-path",
				Aliases: []string{
					"f",
				},
				Usage:     "A path to an app file.",
				Required:  true,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:     "release-note",
				Usage:    "An release note of this revision. Some of services may not support this option.",
				Required: false,
				Sources:  cli.EnvVars(config.ToEnvName("DEPLOYMENT_RELEASE_NOTE")),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.String("name")

			logger.Logger.Info().Msgf("Loading %s config...", name)

			deployment, definition, err := config.CurrentConfig().Deployment(name)

			if err != nil {
				return err
			}

			sourceFilePath := cmd.String("source-path")

			switch deployment.ServiceName {
			case config.DeploygateService:
				dg := deployment.ServiceConfig.(config.DeployGateConfig)

				return task.DeployToDeployGate(ctx, name, dg, sourceFilePath, func(req *service.DeployGateDeployRequest) error {
					if v := cmd.String("release-note"); cmd.IsSet("release-note") {
						req.SetMessage(v)
						req.SetDistributionReleaseNote(v)
					}

					return nil
				})
			case config.LocalService:
				lo := deployment.ServiceConfig.(config.LocalConfig)

				return task.DeployToLocal(ctx, name, lo, sourceFilePath)
			case config.FirebaseAppDistributionService:
				fad := deployment.ServiceConfig.(config.FirebaseAppDistributionConfig)

				return task.DeployToFirebaseAppDistribution(ctx, name, fad, sourceFilePath, func(req *service.FirebaseAppDistributionDeployRequest) error {
					if v := cmd.String("release-note"); cmd.IsSet("release-note") {
						req.SetReleaseNote(v)
					}

					return nil
				})
			case config.TestFlightService:
				tf := deployment.ServiceConfig.(config.TestFlightConfig)

				return task.DeployToTestFlight(ctx, name, tf, sourceFilePath, func(req *service.TestFlightDeployRequest) error {
					return nil
				})
			default:
				custom := deployment.ServiceConfig.(config.CustomServiceConfig)

				return task.DeployToCustomService(ctx, name, deployment.ServiceName, definition, custom, sourceFilePath, func(req *service.CustomServiceDeployRequest) error {
					return nil
				})
			}
		},
	}
}
