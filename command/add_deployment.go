package command

import (
	"github.com/jmatsu/splitter/internal/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

// AddDeploymentConfig command create a template config file for deployment.
func AddDeploymentConfig(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Add a new deployment config to your config file.",
		Description: "This command adds a boilerplate config file of a deployment service to your config file.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "path",
				Usage:    "A path to a config file. This will be used as a new location unless it exists.",
				Required: false,
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "A deployment name.",
			},
			&cli.StringFlag{
				Name:  "service",
				Usage: "A service name.",
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

			name := context.String("name")
			serviceName := context.String("service")

			if name == "" {
				return errors.New("name must be non-empty")
			}

			if serviceName == "" {
				return errors.New("service must be non-empty")
			}

			conf := *config.CurrentConfig()

			if err := conf.AddDeployment(name, serviceName); err != nil {
				return errors.Wrapf(err, "couldn't add a deployment configuration")
			}
			return conf.Dump(path)
		},
	}
}
