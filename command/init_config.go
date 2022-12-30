package command

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func InitConfig(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Initialize your config file",
		Description: "This command generates a boilerplate config file.",
		Flags: []cli.Flag{
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
				return fmt.Errorf("cannot get the current working directory: %v", err)
			} else {
				path = filepath.Join(wd, config.DefaultConfigName)
			}

			if _, err := os.Stat(path); err == nil && !context.Bool("overwrite") {
				return fmt.Errorf("%s already exists. Please add --overwrite option to overwrite the file anyway", path)
			}

			config := config.NewConfig()
			return config.Dump(path)
		},
	}
}
