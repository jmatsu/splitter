package main

import (
	"fmt"
	"github.com/jmatsu/splitter/command"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	version   = "undefined"
	commit    = "undefined"
	timestamp = "undefined" // Dummy
)

func main() {
	app := &cli.App{
		Name:    "splitter",
		Usage:   "An isolated command to distribute your apps to elsewhere",
		Version: fmt.Sprintf("%s (git revision %s) %s", version, commit, timestamp),
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "config",
				Usage:    "A path to a config file.",
				Required: false,
				Action: func(context *cli.Context, s cli.Path) error {
					if _, err := os.Stat(s); err == nil {
						return nil
					} else {
						return fmt.Errorf("%s is not found", s)
					}
				},
				EnvVars: []string{
					config.ToEnvName("CONFIG_FILE"),
				},
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:     "format",
				Usage:    "The output style of command outputs. This may work only for some commands.",
				Required: false,
				Value:    "pretty",
			},
			&cli.BoolFlag{
				Name:     "debug",
				Usage:    "Show debug logs",
				Required: false,
				Value:    false,
				EnvVars: []string{
					config.ToEnvName("DEBUG"),
				},
			},
		},
		Before: func(context *cli.Context) error {
			if context.Bool("debug") {
				logger.SetDebugMode()
			}

			var path *string

			if v := context.Path("config"); context.IsSet("config") {
				path = &v
			}

			if err := config.LoadConfig(path); err != nil {
				return err
			}

			conf := config.GetConfig()

			if newStyle := context.String("format"); context.IsSet("format") || conf.FormatStyle() == "" {
				conf.SetFormatStyle(newStyle)
			}

			if err := format.SetStyle(conf.FormatStyle()); err != nil {
				return err
			}

			return nil
		},
		Commands: []*cli.Command{
			command.InitConfig("init", []string{}),
			command.DeployGate("deploygate", []string{"dg"}),
			command.Local("local", []string{""}),
			command.Distribute("distribute", []string{}),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Logger.Fatal().Err(err).Msg("command exited with non-zero code")
	}
}
