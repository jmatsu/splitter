package task

import (
	"context"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
)

func DistributeToFirebaseAppDistribution(ctx context.Context, conf config.FirebaseAppDistributionConfig, filePath string, builder func(req *service.FirebaseAppDistributionUploadAppRequest)) error {
	if err := conf.Validate(); err != nil {
		return errors.Wrap(err, "the built config is invalid")
	}

	provider := service.NewFirebaseAppDistributionProvider(ctx, &conf)

	formatter := NewFormatter()
	formatter.TableBuilder = FirebaseAppDistributionTableBuilder

	if response, err := provider.Distribute(filePath, builder); err != nil {
		return errors.Wrap(err, "cannot distribute this app")
	} else if err := formatter.Format(response); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	return nil
}

var FirebaseAppDistributionTableBuilder = func(w table.Writer, v any) {
	resp := v.(service.FirebaseAppDistributionDistributionResult)

	w.AppendHeader(table.Row{
		"Key", "Value",
	})

	w.AppendRows([]table.Row{
		{"Firebase App Distribution", ""},
	})
	w.AppendSeparator()

	w.AppendRows([]table.Row{
		{"Processed State", resp.Result},
		{"First Uploaded At", resp.Release.CreatedAt},
		{"First Uploaded At", resp.Release.CreatedAt},
	})

	if resp.Release.ReleaseNote != nil {
		w.AppendRows([]table.Row{
			{"Release Note", resp.Release.ReleaseNote.Text},
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

	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Version Code", resp.Release.BuildVersion},
		{"App Version Name", resp.Release.DisplayVersion},
	})
}
