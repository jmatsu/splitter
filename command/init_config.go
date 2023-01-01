package command

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

// InitConfig command create a template config file.
func InitConfig(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Initialize your config file.",
		Description: "This command generates a boilerplate config file.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "path",
				Usage:    "A path to a new config file.",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "overwrite",
				Usage:    "Allow overriding the existing file if true, otherwise false.",
				Required: false,
				Value:    false,
			},
		},
		Action: func(context *cli.Context) error {
			var path string

			if context.IsSet("path") {
				path = context.String("path")
			} else if wd, err := os.Getwd(); err != nil {
				return errors.Wrap(err, "cannot get the current working directory")
			} else {
				path = filepath.Join(wd, config.DefaultConfigName)
			}

			if _, err := os.Stat(path); err == nil && !context.Bool("overwrite") {
				return errors.New(fmt.Sprintf("%s already exists. Please add --overwrite option to overwrite the file anyway", path))
			}

			conf := config.NewConfig()
			return conf.Dump(path)
		},
		Subcommands: []*cli.Command{},
	}
}
