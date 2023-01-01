package command

import (
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/jmatsu/splitter/task"
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

			return task.DistributeToFirebaseAppDistribution(context.Context, conf, context.String("source-file"), func(req *service.FirebaseAppDistributionUploadAppRequest) {
				if v := context.String("release-note"); context.IsSet("release-note") {
					req.SetReleaseNote(v)
				}
			})
		},
	}
}
