package command

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
	"github.com/urfave/cli/v3"
)

// Flag names are shared by the definitions and the action so that they cannot drift apart.
const (
	deployGateAppOwnerNameFlag            = "app-owner-name"
	deployGateApiTokenFlag                = "api-token"
	deployGateSourcePathFlag              = "source-path"
	deployGateMessageFlag                 = "message"
	deployGateDistributionAccessKeyFlag   = "distribution-access-key"
	deployGateDistributionNameFlag        = "distribution-name"
	deployGateDistributionReleaseNoteFlag = "distribution-release-note"
	deployGateDisableIOSNotificationFlag  = "disable-ios-notification"
)

// DeployGate command distributes your app to DeployGate. This command is standalone so this does not use the values for DeployGate in your config file.
// ref: https://deploygate.com/
func DeployGate(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Deploy your apps to DeployGate.",
		Description: "You can distribute your apps to DeployGate.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: deployGateAppOwnerNameFlag,
				Aliases: []string{
					"n",
				},
				Usage:    "User name or Organization name.",
				Required: true,
				Sources:  cli.EnvVars("DEPLOYGATE_APP_OWNER_NAME"),
			},
			&cli.StringFlag{
				Name: deployGateApiTokenFlag,
				Aliases: []string{
					"t",
				},
				Usage:    "The api token of the app owner.",
				Required: true,
				Sources:  cli.EnvVars("DEPLOYGATE_API_TOKEN"),
			},
			&cli.StringFlag{
				Name: deployGateSourcePathFlag,
				Aliases: []string{
					"f",
				},
				Usage:     "A path to an app file.",
				Required:  true,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name: deployGateMessageFlag,
				Aliases: []string{
					"m",
				},
				Usage:    "A short message of this revision.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     deployGateDistributionAccessKeyFlag,
				Usage:    "An access key of a distribution that must exist. If the both of key and name are specified, key takes priority.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     deployGateDistributionNameFlag,
				Usage:    "An name (title) of a distribution that does not have to exist. If the both of key and name are specified, key takes priority.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     deployGateDistributionReleaseNoteFlag,
				Usage:    "An release note of this revision that will be available only while being distributed via the specified distribution.",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     deployGateDisableIOSNotificationFlag,
				Usage:    "Specify this file if you would like to disable notifications for iOS.",
				Required: false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conf := config.DeployGateConfig{
				AppOwnerName: cmd.String(deployGateAppOwnerNameFlag),
				ApiToken:     cmd.String(deployGateApiTokenFlag),
			}

			return task.DeployToDeployGate(ctx, conf, cmd.String(deployGateSourcePathFlag), func(req *service.DeployGateDeployRequest) error {
				if v := cmd.String(deployGateMessageFlag); cmd.IsSet(deployGateMessageFlag) {
					req.SetMessage(v)
				}

				if v := cmd.String(deployGateDistributionAccessKeyFlag); cmd.IsSet(deployGateDistributionAccessKeyFlag) {
					req.SetDistributionAccessKey(v)
				}

				if v := cmd.String(deployGateDistributionNameFlag); cmd.IsSet(deployGateDistributionNameFlag) {
					req.SetDistributionName(v)
				}

				if v := cmd.String(deployGateDistributionReleaseNoteFlag); cmd.IsSet(deployGateDistributionReleaseNoteFlag) {
					req.SetDistributionReleaseNote(v)
				}

				req.SetIOSDisableNotification(cmd.Bool(deployGateDisableIOSNotificationFlag))

				return nil
			})
		},
	}
}
