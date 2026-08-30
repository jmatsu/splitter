package command

import (
	"context"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
	"github.com/urfave/cli/v3"
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
				Sources:  cli.EnvVars("TESTFLIGHT_APPLE_ID"),
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
				Name: "password",
				Aliases: []string{
					"p",
				},
				Usage:    "App specific password",
				Required: false,
				Sources:  cli.EnvVars("TESTFLIGHT_PASSWORD", "TEST_FLIGHT_PASSWORD"),
			},
			&cli.StringFlag{
				Name:     "api-key",
				Usage:    "API key",
				Required: false,
				Sources:  cli.EnvVars("TESTFLIGHT_API_KEY", "TEST_FLIGHT_API_KEY"),
			},
			&cli.StringFlag{
				Name:     "issuer-id",
				Usage:    "Issuer ID of API Key",
				Required: false,
				Sources:  cli.EnvVars("TESTFLIGHT_ISSUER_ID", "TEST_FLIGHT_ISSUER_ID"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conf := config.TestFlightConfig{
				AppleID:  cmd.String("apple-id"),
				Password: cmd.String("password"),
				ApiKey:   cmd.String("api-key"),
				IssuerID: cmd.String("issuer-id"),
			}

			return task.DeployToTestFlight(ctx, conf, cmd.String("source-path"), func(req *service.TestFlightDeployRequest) error {
				// no-op
				return nil
			})
		},
	}
}
