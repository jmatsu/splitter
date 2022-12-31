package main

import (
	"fmt"
	"github.com/jmatsu/splitter/command"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
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
				Value:    config.DefaultFormat,
				EnvVars: []string{
					config.ToEnvName("FORMAT"),
				},
			},
			&cli.BoolFlag{
				Name:     "async",
				Usage:    "Do not wait for the processing on the provider if awaiting is supported.",
				Required: false,
				Value:    false,
				EnvVars: []string{
					config.ToEnvName("ASYNC"),
				},
			},
			&cli.StringFlag{
				Name:     "log-level",
				Usage:    "Set log level",
				Required: false,
				EnvVars: []string{
					config.ToEnvName("LOG_LEVEL"),
				},
				DefaultText: logger.DefaultLogLevel,
			},
			&cli.StringFlag{
				Name:        "network-timeout",
				Usage:       "Set network timeout for read/connection timeout",
				Required:    false,
				DefaultText: config.DefaultNetworkTimeout,
				EnvVars: []string{
					config.ToEnvName("NETWORK_TIMEOUT"),
				},
			},
			&cli.StringFlag{
				Name:        "wait-timeout",
				Usage:       "Set wait timeout for polling services' processing states",
				Required:    false,
				DefaultText: config.DefaultWaitTimeout,
				EnvVars: []string{
					config.ToEnvName("WAIT_TIMEOUT"),
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

			if v := context.String("format"); context.IsSet("format") {
				conf.SetFormatStyle(v)
			}

			if v := context.String("network-timeout"); context.IsSet("network-timeout") {
				conf.SetNetworkTimeout(v)
			}

			if v := context.String("wait-timeout"); context.IsSet("wait-timeout") {
				conf.SetWaitTimeout(v)
			}

			if v := context.Bool("async"); context.IsSet("async") {
				conf.Async = v
			}

			if err := conf.Validate(); err != nil {
				return errors.Wrap(err, "options contain invalid values or conflict with the current config file")
			}

			net.Configure(conf.NetworkTimeout())
			format.Configure(conf.FormatStyle())

			logger.Logger.Debug().Msgf("format style: %s", conf.FormatStyle())
			logger.Logger.Debug().Msgf("async mode: %t", conf.Async)

			return nil
		},
		Commands: []*cli.Command{
			command.InitConfig("init", []string{}),
			command.Local("local", []string{""}),
			command.DeployGate("deploygate", []string{"dg"}),
			command.FirebaseAppDistribution("firebase-app-distribution", []string{"firebase", "fad"}),
			command.Distribute("distribute", []string{}),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Logger.Trace().Stack().Err(err).Msg("")
		logger.Logger.Fatal().Err(err).Msg("command exited with non-zero code")
	}
}
