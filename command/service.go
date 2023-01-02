package command

import (
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
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
			&cli.PathFlag{
				Name: "source-path",
				Aliases: []string{
					"f",
				},
				Usage:    "A path to an app file.",
				Required: true,
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
				Name:     "name",
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
		Action: func(context *cli.Context) error {
			conf := config.CustomServiceConfig{
				AuthToken: context.String("auth-token"),
			}

			def, err := config.CurrentConfig().Definition(context.String("name"))

			if err != nil {
				return errors.Wrapf(err, "cannot get a definition")
			}

			return task.DeployToCustomService(context.Context, def, conf, context.String("source-path"), func(req *service.CustomServiceDeployRequest) {
				if headers := context.StringSlice("header"); context.IsSet("header") {
					for _, header := range headers {
						name, value, _ := strings.Cut(header, "=")
						req.SetHeader(name, value)
					}
				}
				if params := context.StringSlice("query-param"); context.IsSet("query-param") {
					for _, param := range params {
						name, value, _ := strings.Cut(param, "=")
						if req.HasQueryParam(name) {
							req.AddQueryParam(name, value)
						} else {
							req.SetQueryParam(name, value)
						}
					}
				}
				if params := context.StringSlice("form-param"); context.IsSet("form-param") {
					for _, param := range params {
						name, value, _ := strings.Cut(param, "=")
						req.SetFormParam(name, value)
					}
				}
			})
		},
	}
}
