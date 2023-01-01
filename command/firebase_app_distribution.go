package command

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/provider/firebase_app_distribution"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// FirebaseAppDistribution command distributes your app to Firebase App Distribution. This command is standalone so this does not use the values for Firebase App Distribution in your config file.
// ref: https://firebase.google.com/docs/app-distribution
func FirebaseAppDistribution(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Distribute your apps to Firebase App Distribution",
		Description: "You can distribute your apps to Firebase App Distribution. Please note that this command does not respect for static config files. All parameters have to be specified from command line options.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "app-id",
				Usage:    "Firebase App ID e.g. 1:123456578:android:xxxxxxx",
				Required: true,
				EnvVars:  []string{"FIREBASE_APP_ID"},
			},
			&cli.PathFlag{
				Name: "source-file",
				Aliases: []string{
					"f",
				},
				Usage:    "A path to an app file",
				Required: true,
			},
			&cli.StringFlag{
				Name: "access-token",
				Aliases: []string{
					"t",
				},
				Usage:    "The access token to use for this distribution",
				Required: false,
				EnvVars:  []string{"FIREBASE_CLI_TOKEN"},
			},
			&cli.PathFlag{
				Name: "credentials",
				Aliases: []string{
					"c",
				},
				Usage:    "A path to a credentials json file",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "release-note",
				Usage:    "An release note of this revision",
				Required: false,
			},
		},
		Action: func(context *cli.Context) error {
			conf := config.FirebaseAppDistributionConfig{
				AccessToken:           context.String("access-token"),
				GoogleCredentialsPath: context.String("credentials"),
				AppId:                 context.String("app-id"),
			}

			if err := conf.Validate(); err != nil {
				return errors.Wrap(err, "given flags may be insufficient or invalid")
			}

			return distributeFirebaseAppDistribution(context.Context, &conf, context.String("source-file"), func(req *firebase_app_distribution.FirebaseAppDistributionUploadAppRequest) {
				if v := context.String("release-note"); context.IsSet("release-note") {
					req.SetReleaseNote(v)
				}
			})
		},
	}
}

func distributeFirebaseAppDistribution(ctx context.Context, conf *config.FirebaseAppDistributionConfig, filePath string, builder func(req *firebase_app_distribution.FirebaseAppDistributionUploadAppRequest)) error {
	provider := firebase_app_distribution.NewFirebaseAppDistributionProvider(ctx, conf)

	if response, err := provider.Distribute(filePath, builder); err != nil {
		return err
	} else if format.IsRaw() {
		fmt.Println(response.RawJson)
	} else if err := format.Format(*response, firebase_app_distribution.FirebaseAppDistributionTableBuilder); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	return nil
}
