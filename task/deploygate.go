package task

import (
	"context"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
	"strings"
)

func DeployToDeployGate(ctx context.Context, conf config.DeployGateConfig, filePath string, builder func(req *service.DeployGateDeployRequest) error) error {
	if err := conf.Validate(); err != nil {
		return errors.Wrap(err, "the built config is invalid")
	}

	provider := service.NewDeployGateProvider(ctx, &conf)

	formatter := NewFormatter()
	formatter.TableBuilder = deployGateTableBuilder

	if response, err := provider.Deploy(filePath, builder); err != nil {
		return errors.Wrap(err, "cannot deploy this app")
	} else if err := formatter.Format(response); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	return nil
}

var deployGateTableBuilder = func(w table.Writer, v any) {
	resp := v.(service.DeployGateDeployResult).Results

	w.AppendHeader(table.Row{
		"Key", "Value",
	})

	w.AppendRows([]table.Row{
		{"DeployGate Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"Owner Name", resp.User.Name},
		{"Revision", resp.Revision},
	})

	switch strings.ToLower(resp.OsName) {
	case "android":
		w.AppendRows([]table.Row{
			{"APK Download URL", resp.DownloadUrl},
		})
	case "ios":
		w.AppendRows([]table.Row{
			{"IPA Download URL", resp.DownloadUrl},
		})
	}

	if resp.Distribution != nil {
		w.AppendRows([]table.Row{
			{"Deployment Name", resp.Distribution.Title},
			{"Deployment AccessKey", resp.Distribution.AccessKey},
			{"Deployment URL", resp.Distribution.Url},
		})
	}

	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"App Property", ""},
	})
	w.AppendSeparator()

	switch strings.ToLower(resp.OsName) {
	case "android":
		w.AppendRows([]table.Row{
			{"Display Name", resp.Name},
			{"Package Name", resp.PackageName},
			{"Version Code", resp.VersionCode},
			{"Version Name", resp.VersionName},
			{"Min SDK Version", resp.SdkVersion},
			{"Target SDK Version", resp.TargetSdkVersion},
		})

	case "ios":
		w.AppendRows([]table.Row{
			{"Display Name", resp.Name},
			{"Bundle Identifier", resp.PackageName},
			{"Version Code", resp.VersionCode},
			{"Version Name", resp.VersionName},
			{"Build SDK Version", resp.RawSdkVersion},
		})
	}
}
