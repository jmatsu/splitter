package task

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/service"
	"os"
)

type TableBuilder = func(writer table.Writer, r any)

type Formatter struct {
	style        config.FormatStyle
	TableBuilder TableBuilder
}

func NewFormatter() *Formatter {
	return &Formatter{
		style: config.CurrentConfig().FormatStyle(),
	}
}

func (f *Formatter) Format(r service.DistributionResult) error {
	w := table.NewWriter()
	w.SetOutputMirror(os.Stdout)
	w.SetStyle(table.StyleDefault)

	if f.TableBuilder == nil && f.style != config.RawFormat {
		logger.Logger.Error().Msg("pretty formatter is not found so fall back into a raw format")
		f.style = config.RawFormat
	}

	switch f.style {
	case config.RawFormat:
		fmt.Println(r.RawJsonResponse())
	case config.PrettyFormat:
		f.TableBuilder(w, r.ValueResponse())

		w.Render()
	case config.MarkdownFormat:
		f.TableBuilder(w, r.ValueResponse())
		w.RenderMarkdown()
	}

	return nil
}
