package command

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"strings"
)

// CustomService command distributes your app to the defined service in the config file.
func CustomService(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Deploy your apps to the defined service in the config file.",
		Description: "You can distribute your apps to the defined service in the config file.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "source-path",
				Aliases: []string{
					"f",
				},
				Usage:     "A path to an app file.",
				Required:  true,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name: "auth-token",
				Aliases: []string{
					"t",
				},
				Usage:    "The auth token to use for this distribution.",
				Required: true,
			},
			&cli.StringFlag{
				Name: "name",
				Aliases: []string{
					"n",
				},
				Usage:    "A service name in the config file.",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "header",
				Usage:    "Append <key>=<value> to headers",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:     "query-param",
				Usage:    "Append <key>=<value> to query parameters",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:     "form-param",
				Usage:    "Append <key>=<value> to form parameters",
				Required: false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conf := config.CustomServiceConfig{
				AuthToken: cmd.String("auth-token"),
			}

			def, err := config.CurrentConfig().Definition(cmd.String("name"))

			if err != nil {
				return errors.Wrapf(err, "cannot get a definition")
			}

			return task.DeployToCustomService(ctx, def, conf, cmd.String("source-path"), func(req *service.CustomServiceDeployRequest) error {
				if headers := cmd.StringSlice("header"); cmd.IsSet("header") {
					for _, header := range headers {
						if name, value, ok := strings.Cut(header, "="); ok {
							req.SetHeader(name, value)
						} else {
							return errors.New(fmt.Sprintf("--header %s must follow <name>=<value> format", header))
						}
					}
				}
				if params := cmd.StringSlice("query-param"); cmd.IsSet("query-param") {
					for _, param := range params {
						if name, value, ok := strings.Cut(param, "="); ok {
							if req.HasQueryParam(name) {
								req.AddQueryParam(name, value)
							} else {
								req.SetQueryParam(name, value)
							}
						} else {
							return errors.New(fmt.Sprintf("--query-param %s must follow <name>=<value> format", param))
						}
					}
				}
				if params := cmd.StringSlice("form-param"); cmd.IsSet("form-param") {
					for _, param := range params {
						if name, value, ok := strings.Cut(param, "="); ok {
							req.SetFormParam(name, value)
						} else {
							return errors.New(fmt.Sprintf("--form-param %s must follow <name>=<value> format", param))
						}
					}
				}

				return nil
			})
		},
	}
}
