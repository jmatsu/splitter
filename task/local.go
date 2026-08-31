package task

import (
	"context"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
)

func DeployToLocal(ctx context.Context, deploymentName string, conf config.LocalConfig, filePath string) error {
	if err := conf.Validate(); err != nil {
		return errors.Wrap(err, "the built config is invalid")
	}

	provider := service.NewLocalProvider(ctx, &conf)

	formatter := NewFormatter()
	formatter.TableBuilder = localTableBuilder

	dumper := NewDumper(config.LocalService, deploymentName, filePath)

	response, err := provider.Deploy(filePath)

	if err != nil {
		return errors.Wrap(err, "cannot deploy this app")
	}

	if err := formatter.Format(response); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	if err := dumper.Dump(response); err != nil {
		return errors.Wrap(err, "cannot dump the response")
	}

	return nil
}

var localTableBuilder = func(w table.Writer, v any) {
	resp := v.(service.LocalDeployResult)

	w.AppendHeader(table.Row{
		"Key", "Value",
	})

	w.AppendRows([]table.Row{
		{"Source Path", resp.SourceFilePath},
		{"Destination Path", resp.DestinationFilePath},
	})

	w.AppendRows([]table.Row{
		{"SideEffect", resp.SideEffect},
	})
}
