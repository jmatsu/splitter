package main

import (
	"fmt"
	"github.com/jmatsu/splitter/command"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal"
	"github.com/jmatsu/splitter/internal/logger"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "splitter",
		Usage: "An isolated command to distribute your apps to elsewhere",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "config",
				Required: false,
				Action: func(context *cli.Context, s cli.Path) error {
					if _, err := os.Stat(s); err == nil {
						return nil
					} else {
						return fmt.Errorf("%s is not found", s)
					}
				},
				EnvVars: []string{
					internal.ToEnvName("CONFIG_FILE"),
				},
			},
			&cli.StringFlag{
				Name:     "format",
				Usage:    "Print command outputs by following the specified style. This may work only for some commands.",
				Required: false,
				Value:    "pretty",
			},
			&cli.BoolFlag{
				Name:     "debug",
				Usage:    "Show debug logs",
				Required: false,
				Value:    false,
				EnvVars: []string{
					internal.ToEnvName("DEBUG"),
				},
				Action: func(context *cli.Context, b bool) error {
					if b {
						logger.SetDebugMode()
					}

					return nil
				},
			},
		},
		Before: func(context *cli.Context) error {
			var path *string

			if v := context.Path("path"); context.IsSet("path") {
				path = &v
			}

			if err := internal.LoadConfig(path); err != nil {
				return err
			}

			config := internal.GetConfig()

			if newStyle := context.String("format"); context.IsSet("format") || config.FormatStyle() == "" {
				config.SetFormatStyle(newStyle)
			}

			if err := format.SetStyle(config.FormatStyle()); err != nil {
				return err
			}

			return nil
		},
		Commands: []*cli.Command{
			command.InitConfig("init", []string{}),
			command.DeployGate("deploygate", []string{"dg"}),
			{
				Name:        "distribute",
				Subcommands: []*cli.Command{},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Logger.Fatal().Err(err).Msg("command exited with non-zero code")
	}
}
