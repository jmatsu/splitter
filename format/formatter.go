package format

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"os"
	"reflect"
)

var currentStyle config.FormatStyle

type TableBuilder = func(writer table.Writer, v any)

func SetStyle(style config.FormatStyle) {
	currentStyle = style

	logger.Logger.Debug().
		Str("style", style).
		Msg("Configuring formatter")
}

func IsRaw() bool {
	return currentStyle == config.RawFormat
}

func Format(v any, tableBuilder TableBuilder) error {
	if reflect.ValueOf(v).Kind() != reflect.Struct {
		panic("v must be struct")
	}

	w := table.NewWriter()
	w.SetOutputMirror(os.Stdout)

	tableBuilder(w, v)

	switch currentStyle {
	case config.RawFormat:
		panic("call fmt.Printf directly in advance")
	case config.PrettyFormat:
		w.SetStyle(table.StyleDefault)

		w.Render()
	case config.MarkdownFormat:
		w.RenderMarkdown()
	}

	return nil
}
