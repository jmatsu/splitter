package command

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/provider/deploygate"
	"github.com/jmatsu/splitter/provider/firebase_app_distribution"
	"github.com/urfave/cli/v2"
)

func Distribute(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Distribute your apps with following the configuration",
		Description: "You can distribute your apps to supported services based on pre-defined service configuration.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "name",
				Aliases: []string{
					"n",
				},
				Usage:    "distribution name in your configuration file",
				Required: true,
				EnvVars:  []string{config.ToEnvName("DISTRIBUTION_NAME")},
			},
			&cli.PathFlag{
				Name: "source-file",
				Aliases: []string{
					"f",
				},
				Usage:    "A path to an app file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "release-note",
				Usage:    "An release note of this revision. Some of services may not support this option",
				Required: false,
				EnvVars:  []string{config.ToEnvName("DISTRIBUTION_RELEASE_NOTE")},
			},
		},
		Action: func(context *cli.Context) error {
			conf := config.GetConfig()

			name := context.String("name")

			d, err := conf.GetDistribution(name)

			if err != nil {
				return err
			}

			sourceFilePath := context.String("source-file")

			switch d.ServiceName {
			case config.DeploygateService:
				dg := d.ServiceConfig.(*config.DeployGateConfig)

				return distributeDeployGate(context.Context, dg, sourceFilePath, func(req *deploygate.UploadRequest) {
					if v := context.String("release-note"); context.IsSet("release-note") {
						req.SetMessage(v)
						req.SetDistributionReleaseNote(v)
					}
				})
			case config.LocalService:
				lo := d.ServiceConfig.(*config.LocalConfig)

				return distributeLocal(context.Context, lo, sourceFilePath)
			case config.FirebaseAppDistributionService:
				fad := d.ServiceConfig.(*config.FirebaseAppDistributionConfig)

				return distributeFirebaseAppDistribution(context.Context, fad, sourceFilePath, func(req *firebase_app_distribution.UploadRequest) {
					if v := context.String("release-note"); context.IsSet("release-note") {
						req.SetReleaseNote(v)
					}
				})
			default:
				return fmt.Errorf("%s is not implemented yet", d.ServiceName)
			}
		},
	}
}
