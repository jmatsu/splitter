package command

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/provider/local"
	"github.com/pkg/errors"
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

			if err := conf.Validate(); err != nil {
				return errors.Wrap(err, "given flags may be insufficient or invalid")
			}

			return distributeLocal(context.Context, &conf, context.String("source-file"))
		},
	}
}

func distributeLocal(ctx context.Context, conf *config.LocalConfig, filePath string) error {
	provider := local.NewProvider(ctx, conf)

	if response, err := provider.Distribute(filePath); err != nil {
		return err
	} else if format.IsRaw() {
		fmt.Println(response.RawJson)
	} else if err := format.Format(*response, local.TableBuilder); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	return nil
}
