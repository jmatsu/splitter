package command

import (
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
	"github.com/urfave/cli/v2"
)

// TestFlight command distributes your app to TestFlight. This command is standalone so this does not use the values for TestFlight in your config file.
func TestFlight(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Deploy your apps to TestFlight.",
		Description: "You can distribute your apps to TestFlight.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "apple-id",
				Usage:    "Your AppleID",
				Required: true,
				EnvVars:  []string{"TESTFLIGHT_APPLE_ID"},
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
				Name: "password",
				Aliases: []string{
					"p",
				},
				Usage:    "App specific password",
				Required: false,
				EnvVars:  []string{"TESTFLIGHT_PASSWORD"},
			},
		},
		Action: func(context *cli.Context) error {
			conf := config.TestFlightConfig{
				AppleID:  context.String("apple-id"),
				Password: context.String("password"),
			}

			return task.DeployToTestFlight(context.Context, conf, context.String("source-path"), func(req *service.TestFlightDeployRequest) error {
				// no-op
				return nil
			})
		},
	}
}
