package task

import (
	"context"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/service"
	"github.com/pkg/errors"
)

func DeployToTestFlight(ctx context.Context, conf config.TestFlightConfig, filePath string, builder func(req *service.TestFlightDeployRequest) error) error {
	if err := conf.Validate(); err != nil {
		return errors.Wrap(err, "the built config is invalid")
	}

	provider := service.NewTestFlightProvider(ctx, &conf)

	formatter := NewFormatter()
	formatter.TableBuilder = testFlightTableBuilder

	if response, err := provider.Deploy(filePath, builder); err != nil {
		return errors.Wrap(err, "cannot deploy this app")
	} else if err := formatter.Format(response); err != nil {
		return errors.Wrap(err, "cannot format the response")
	}

	return nil
}

var testFlightTableBuilder = func(w table.Writer, v any) {
	resp := v.(service.TestFlightDeployResult)

	w.AppendHeader(table.Row{
		"Key", "Value",
	})

	w.AppendRows([]table.Row{
		{"TestFlight Property", ""},
	})
	w.AppendSeparator()
	w.AppendRows([]table.Row{
		{"Result", resp.SuccessMessage},
		{"altool Version", resp.ToolVersion},
	})
}
