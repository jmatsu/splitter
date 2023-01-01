package task

import (
	"context"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
)

func DistributeToLocal(ctx context.Context, conf config.LocalConfig, filePath string) error {
	if err := conf.Validate(); err != nil {
		return errors.Wrap(err, "the built config is invalid")
	}

	provider := service.NewLocalProvider(ctx, &conf)

	formatter := NewFormatter()
	formatter.TableBuilder = localTableBuilder

	if response, err := provider.Distribute(filePath); err != nil {
		return errors.Wrap(err, "cannot distribute this app")
	} else if err := formatter.Format(response); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	return nil
}

var localTableBuilder = func(w table.Writer, v any) {
	resp := v.(service.LocalDistributionResult)

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
