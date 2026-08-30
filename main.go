package main

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/command"
	"github.com/jmatsu/splitter/internal"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"os"
)

func main() {
	cmd := &cli.Command{
		Name:      "splitter",
		Usage:     "A command to deploy your apps to several mobile app distribution services.",
		Version:   fmt.Sprintf("%s (git revision %s)", internal.Version, internal.Commit),
		Copyright: "Jumpei Matsuda (@jmatsu)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "A path to a config file.",
				Required: false,
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					if _, err := os.Stat(s); err == nil {
						return nil
					} else {
						return errors.New(fmt.Sprintf("%s is not found", s))
					}
				},
				Sources:   cli.EnvVars(config.ToEnvName("CONFIG_FILE")),
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:     "format",
				Usage:    "The output style of command outputs.",
				Required: false,
				Value:    config.DefaultFormat,
				Sources:  cli.EnvVars(config.ToEnvName("FORMAT")),
			},
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "Set log level.",
				Required:    false,
				Sources:     cli.EnvVars(config.ToEnvName("LOG_LEVEL")),
				DefaultText: logger.DefaultLogLevel,
			},
			&cli.StringFlag{
				Name:        "network-timeout",
				Usage:       "Set network timeout for read/connection timeout.",
				Required:    false,
				DefaultText: config.DefaultNetworkTimeout,
				Sources:     cli.EnvVars(config.ToEnvName("NETWORK_TIMEOUT")),
			},
			&cli.StringFlag{
				Name:        "wait-timeout",
				Usage:       "Set wait timeout for polling services' processing states.",
				Required:    false,
				DefaultText: config.DefaultWaitTimeout,
				Sources:     cli.EnvVars(config.ToEnvName("WAIT_TIMEOUT")),
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if logLevel := cmd.String("log-level"); cmd.IsSet("log-level") {
				logger.SetLogLevel(logLevel)
			}

			var path *string

			if v := cmd.String("config"); cmd.IsSet("config") {
				path = &v
			}

			if err := config.LoadGlobalConfig(path); err != nil {
				return ctx, err
			}

			if v := cmd.String("format"); cmd.IsSet("format") {
				config.SetGlobalFormatStyle(v)
			}

			if v := cmd.String("network-timeout"); cmd.IsSet("network-timeout") {
				config.SetGlobalNetworkTimeout(v)
			}

			if v := cmd.String("wait-timeout"); cmd.IsSet("wait-timeout") {
				config.SetGlobalWaitTimeout(v)
			}

			c := config.CurrentConfig()

			if err := c.Validate(); err != nil {
				return ctx, errors.Wrap(err, "options contain invalid values or conflict with the current config file")
			}

			logger.Logger.Debug().
				Str("network-timeout", c.NetworkTimeout().String()).
				Str("wait-timeout", c.WaitTimeout().String()).
				Str("format-style", c.FormatStyle()).
				Msg("configuration has been initialized")

			return ctx, nil
		},
		Commands: []*cli.Command{
			command.InitConfig("init", []string{}),
			command.Local("local", []string{""}),
			command.DeployGate("deploygate", []string{"dg"}),
			command.FirebaseAppDistribution("firebase-app-distribution", []string{"firebase", "fad"}),
			command.Deploy("deploy", []string{}),
			command.AddDeploymentConfig("add-deployment", []string{}),
			command.CustomService("service", []string{}),
			command.TestFlight("test-flight", []string{"tf"}),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logger.Logger.Trace().Stack().Err(err).Msg("")
		logger.Logger.Fatal().Err(err).Msg("command exited with non-zero code")
	}
}
