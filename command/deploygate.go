package command

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/provider/deploygate"
	"github.com/urfave/cli/v2"
)

func DeployGate(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Distribute your apps to DeployGate",
		Description: "You can distribute your apps to DeployGate. Please note that this command does not respect for static config files. All parameters have to be specified from command line options. ref: https://docs.deploygate.com/docs/api/application/upload",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "app-owner-name",
				Aliases: []string{
					"n",
				},
				Usage:    "User name or Organization name",
				Required: true,
				EnvVars:  []string{"DEPLOYGATE_APP_OWNER_NAME"},
			},
			&cli.StringFlag{
				Name: "api-token",
				Aliases: []string{
					"t",
				},
				Usage:    "The api token of the app owner",
				Required: true,
				EnvVars:  []string{"DEPLOYGATE_API_TOKEN"},
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
				Name: "message",
				Aliases: []string{
					"m",
				},
				Usage:    "A short message of this revision",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "distribution-access-key",
				Usage:    "An access key of a distribution that must exist. If the both of key and name are specified, key takes priority.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "distribution-name",
				Usage:    "An name (title) of a distribution that does not have to exist. If the both of key and name are specified, key takes priority.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "distribution-release-note",
				Usage:    "An release note of this revision that will be available only while being distributed via the specified distribution.",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "disable-ios-notification",
				Usage:    "Specify this file if you would like to disable notifications for iOS.",
				Required: false,
			},
		},
		Action: func(context *cli.Context) error {
			conf := config.DeployGateConfig{
				AppOwnerName: context.String("app-owner-name"),
				ApiToken:     context.String("api-token"),
			}

			if err := conf.Validate(); err != nil {
				return fmt.Errorf("given flags may be insufficient or invalid: %v", err)
			}

			return distributeDeployGate(context.Context, &conf, context.String("file"), func(req *deploygate.UploadRequest) {
				if v := context.String("message"); context.IsSet("message") {
					req.SetMessage(v)
				}

				if v := context.String("distribution-key"); context.IsSet("distribution-key") {
					req.SetDistributionAccessKey(v)
				}

				if v := context.String("distribution-name"); context.IsSet("distribution-name") {
					req.SetDistributionName(v)
				}

				if v := context.String("release-note"); context.IsSet("release-note") {
					req.SetDistributionReleaseNote(v)
				}

				req.SetIOSDisableNotification(context.Bool("disable-ios-notification"))
			})
		},
	}
}

func distributeDeployGate(ctx context.Context, conf *config.DeployGateConfig, filePath string, builder func(req *deploygate.UploadRequest)) error {
	provider := deploygate.NewProvider(ctx, conf)

	if response, err := provider.Distribute(filePath, builder); err != nil {
		return err
	} else if format.IsRaw() {
		fmt.Println(response.RawJson)
	} else if err := format.Format(response, deploygate.TableBuilder); err != nil {
		return fmt.Errorf("cannot format the response: %v", err)
	}

	return nil
}
