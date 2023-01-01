package command

import (
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/task"
	"github.com/urfave/cli/v2"
	"os"
)

// Local command copy/move your app to another location. This command is standalone so this does not use the values for Local in your config file.
func Local(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Move/Copy your apps to another location",
		Description: "You can manage/copy your apps to another location. Please note that this command does not respect for static config files.",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name: "source-file",
				Aliases: []string{
					"f",
				},
				Usage:    "A source path to an app file.",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "destination",
				Usage:    "A destination path to an app file.",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "delete-source",
				Usage:    "Specify true if you would not like to keep the source file.",
				Required: false,
				Value:    false,
			},
			&cli.BoolFlag{
				Name:     "overwrite",
				Usage:    "Specify true if you allow to overwrite the existing destination file.",
				Required: false,
				Value:    false,
			},
			&cli.UintFlag{
				Name:        "mode",
				Usage:       "The final file mode of the destination path.",
				Required:    false,
				Value:       0,
				DefaultText: "The same to the source",
			},
		},
		Action: func(context *cli.Context) error {
			conf := config.LocalConfig{
				DestinationPath: context.String("destination"),
				DeleteSource:    context.Bool("delete-source"),
				AllowOverwrite:  context.Bool("overwrite"),
				FileMode:        os.FileMode(context.Uint("mode")),
			}

			return task.DistributeToLocal(context.Context, conf, context.String("source-file"))
		},
	}
}
