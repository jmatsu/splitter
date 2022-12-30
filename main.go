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
			&cli.StringFlag{
				Name:     "config",
				Required: false,
				Action: func(context *cli.Context, s string) error {
					if _, err := os.Stat(s); err == nil {
						return nil
					} else {
						return fmt.Errorf("%s is not found", s)
					}
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
			},
		},
		Before: func(context *cli.Context) error {
			if context.Bool("debug") {
				logger.SetDebugMode()
			}

			if context.IsSet("path") {
				path := context.String("path")
				if err := internal.LoadConfig(&path); err != nil {
					return err
				}
			} else {
				if err := internal.LoadConfig(nil); err != nil {
					return err
				}
			}

			config := internal.GetConfig()

			if config.Debug() {
				logger.SetDebugMode()
			}

			style := context.String("format")

			if newStyle, ok := config.FormatStyle(); !context.IsSet("format") && ok {
				style = newStyle
			}

			if err := format.SetStyle(style); err != nil {
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
