package command

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/format"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/provider/deploygate"
	"github.com/jmatsu/splitter/provider/firebase_app_distribution"
	"github.com/urfave/cli/v2"
	"strings"
)

func FirebaseAppDistribution(name string, aliases []string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Distribute your apps to Firebase App Distribution",
		Description: "You can distribute your apps to Firebase App Distribution. Please note that this command does not respect for static config files. All parameters have to be specified from command line options.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "project-number",
				Aliases: []string{
					"n",
				},
				Usage:    "Firebase Project Number (not ID)",
				Required: true,
				EnvVars:  []string{"FIREBASE_PROJECT_NUMBER"},
			},
			&cli.StringFlag{
				Name: "access-token",
				Aliases: []string{
					"t",
				},
				Usage:    "The access token to use for this distribution",
				Required: true,
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
				Name:     "os",
				Usage:    "OS name of an app: android or ios (case-insensitive)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "package-name",
				Usage:    "A package name of an app to distribute.",
				Required: true,
			},
		},
		Action: func(context *cli.Context) error {
			conf := config.FirebaseAppDistributionConfig{
				AccessToken:   context.String("access-token"),
				ProjectNumber: context.String("project-number"),
				OsName:        strings.ToLower(context.String("os")),
				PackageName:   context.String("package-name"),
			}

			if err := conf.Validate(); err != nil {
				return fmt.Errorf("given flags may be insufficient or invalid: %v", err)
			}

			return distributeFirebaseAppDistribution(context.Context, &conf, context.String("source-file"), func(req *firebase_app_distribution.UploadRequest) {

			})
		},
	}
}

func distributeFirebaseAppDistribution(ctx context.Context, conf *config.FirebaseAppDistributionConfig, filePath string, builder func(req *firebase_app_distribution.UploadRequest)) error {
	provider := firebase_app_distribution.NewProvider(ctx, conf)

	if response, err := provider.Distribute(filePath, builder); err != nil {
		return err
	} else if format.IsRaw() {
		fmt.Println(response.RawJson)
	} else if err := format.Format(response, deploygate.TableBuilder); err != nil {
		return fmt.Errorf("cannot format the response: %v", err)
	}

	return nil
}
