package command

import (
	"fmt"
	"github.com/jmatsu/splitter/internal"
	"github.com/jmatsu/splitter/provider/deploygate"
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
				EnvVars:  []string{internal.ToEnvName("DISTRIBUTION_NAME")},
			},
			&cli.PathFlag{
				Name: "file",
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
				EnvVars:  []string{internal.ToEnvName("DISTRIBUTION_RELEASE_NOTE")},
			},
		},
		Action: func(context *cli.Context) error {
			config := internal.GetConfig()

			name := context.String("name")

			d, err := config.GetDistribution(name)

			if err != nil {
				return err
			}

			switch d.ServiceName {
			case internal.DeploygateService:
				dg := d.ServiceConfig.(*internal.DeployGateConfig)

				return distributeDeployGate(context.Context, dg, context.String("file"), func(req *deploygate.UploadRequest) {
					if v := context.String("release-note"); context.IsSet("release-note") {
						req.Message = &v
					}
				})
			default:
				return fmt.Errorf("%s is not implemented yet", d.ServiceName)
			}
		},
	}
}
