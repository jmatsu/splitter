package task

import (
	"context"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
)

func DeployToFirebaseAppDistribution(ctx context.Context, conf config.FirebaseAppDistributionConfig, filePath string, builder func(req *service.FirebaseAppDistributionDeployRequest)) error {
	if err := conf.Validate(); err != nil {
		return errors.Wrap(err, "the built config is invalid")
	}

	provider := service.NewFirebaseAppDistributionProvider(ctx, &conf)

	formatter := NewFormatter()
	formatter.TableBuilder = firebaseAppDistributionTableBuilder

	if response, err := provider.Deploy(filePath, builder); err != nil {
		return errors.Wrap(err, "cannot deploy this app")
	} else if err := formatter.Format(response); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	return nil
}

var firebaseAppDistributionTableBuilder = func(w table.Writer, v any) {
	resp := v.(service.FirebaseAppDistributionDeployResult)

	if resp.Response == nil {
		w.SetTitle("The results cannot be rendered for an unknown reason. Please create an issue at https://github.com/jmatsu/splitter/issues")
		return
	}

	release := resp.Response.Release

	w.AppendHeader(table.Row{
		"Key", "Value",
	})

	w.AppendRows([]table.Row{
		{"Firebase App Deployment", ""},
	})
	w.AppendSeparator()

	w.AppendRows([]table.Row{
		{"Processed State", resp.Response.Result},
		{"First Uploaded At", release.CreatedAt},
		{"First Uploaded At", release.CreatedAt},
	})

	if release.ReleaseNote != nil {
		w.AppendRows([]table.Row{
			{"Release Note", release.ReleaseNote.Text},
		})
	}

	// it's okay to use aabInfo != nil as *if android*
	if aabInfo := resp.AabInfo; aabInfo != nil {
		w.AppendRows([]table.Row{
			{"App Bundle Available", aabInfo.Available()},
			{"Play Store Integration", aabInfo.IntegrationState},
		})

		if certificate := aabInfo.TestCertificate; certificate != nil {
			w.AppendRows([]table.Row{
				{"App Bundle Certificate MD5", certificate.Md5},
				{"App Bundle Certificate SHA1", certificate.Sha1},
				{"App Bundle Certificate SHA256", certificate.Sha256},
			})
		}
	}

	w.AppendRows([]table.Row{
		{"Groups", fmt.Sprintf("%d groups", len(resp.GroupAliases))},
		{"Individual Testers", fmt.Sprintf("%d testers", len(resp.TesterEmails))},
	})

	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Version Code", release.BuildVersion},
		{"App Version Name", release.DisplayVersion},
	})
}
