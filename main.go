package main

import (
	"fmt"
	"github.com/jmatsu/splitter/internal"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	var configPath *string

	app := &cli.App{
		Name:  "splitter",
		Usage: "An isolated command to distribute your apps to elsewhere",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Required: false,
				Action: func(context *cli.Context, s string) error {
					if _, err := os.Stat(s); err == nil {
						configPath = &s
						return nil
					} else {
						return fmt.Errorf("%v", err)
					}
				},
			},
		},
		Before: func(context *cli.Context) error {
			return internal.LoadConfig(configPath)
		},
		Commands: []*cli.Command{
			&cli.Command{},
		},
	}

	if err := app.Run(os.Args); err != nil {
		internal.Logger.Fatal().Err(err).Msg("command exited with non-zero code")
	}
}
