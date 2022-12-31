package main

import (
	"fmt"
	"github.com/jmatsu/splitter/command"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
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
						return errors.New(fmt.Sprintf("%s is not found", s))
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
				Name:     "async",
				Usage:    "Do not wait for the processing on the provider if awaiting is supported.",
				Required: false,
				Value:    false,
			},
			&cli.StringFlag{
				Name:     "log-level",
				Usage:    "Set log level",
				Required: false,
				EnvVars: []string{
					config.ToEnvName("LOG_LEVEL"),
				},
			},
		},
		Before: func(context *cli.Context) error {
			if logLevel := context.String("log-level"); context.IsSet("log-level") {
				logger.SetLogLevel(logLevel)
			}

			var path *string

			if v := context.Path("config"); context.IsSet("config") {
				path = &v
			}

			if err := config.LoadGlobalConfig(path); err != nil {
				return err
			}

			conf := config.GetGlobalConfig()

			if newStyle := context.String("format"); context.IsSet("format") || conf.FormatStyle() == "" {
				conf.SetFormatStyle(newStyle)
			}

			if async := context.Bool("async"); context.IsSet("async") {
				conf.Async = async
			}

			if err := format.SetStyle(conf.FormatStyle()); err != nil {
				return err
			}

			logger.Logger.Debug().Msgf("format style: %s", conf.FormatStyle())
			logger.Logger.Debug().Msgf("async mode: %t", conf.Async)

			return nil
		},
		Commands: []*cli.Command{
			command.InitConfig("init", []string{}),
			command.DeployGate("deploygate", []string{"dg"}),
			command.Local("local", []string{""}),
			command.FirebaseAppDistribution("firebase-app-distribution", []string{"firebase", "fad"}),
			command.Distribute("distribute", []string{}),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Logger.Trace().Stack().Err(err).Msg("")
		logger.Logger.Fatal().Err(err).Msg("command exited with non-zero code")
	}
}
