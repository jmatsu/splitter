package format

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/exp/slices"
	"os"
	"reflect"
)

type FormatStyle = string

const (
	Pretty   FormatStyle = "pretty"
	Raw      FormatStyle = "raw"
	Markdown FormatStyle = "markdown"
)

var styles = []FormatStyle{
	Pretty,
	Raw,
	Markdown,
}

var currentStyle FormatStyle

type TableBuilder = func(writer table.Writer, v any)

func SetStyle(style FormatStyle) error {
	if !slices.Contains(styles, style) {
		return fmt.Errorf("%s is unknown style", style)
	}

	currentStyle = style

	return nil
}

func IsRaw() bool {
	return currentStyle == Raw
}

func Format(v any, tableBuilder TableBuilder) error {
	if reflect.ValueOf(v).Kind() != reflect.Struct {
		panic(fmt.Errorf("v must be struct"))
	}

	w := table.NewWriter()
	w.SetOutputMirror(os.Stdout)

	tableBuilder(w, v)

	switch currentStyle {
	case Raw:
		panic(fmt.Errorf("call fmt.Printf directly in advance"))
	case Pretty:
		w.SetStyle(table.StyleDefault)

		w.Render()
	case Markdown:
		w.RenderMarkdown()
	}

	return nil
}
