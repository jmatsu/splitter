package format

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/exp/slices"
	"os"
	"reflect"
)

type FormatStyle = string

const (
	Pretty   FormatStyle = "pretty"
	Json     FormatStyle = "json"
	Markdown FormatStyle = "markdown"
)

var styles = []FormatStyle{
	Pretty,
	Json,
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

func IsJson() bool {
	return currentStyle == Json
}

func Format(v any, tableBuilder TableBuilder) error {
	vRef := reflect.ValueOf(v)

	if vRef.Kind() != reflect.Struct {
		panic(fmt.Errorf("v must be struct"))
	}

	w := table.NewWriter()
	w.SetOutputMirror(os.Stdout)

	switch currentStyle {
	case Json:
		if bytes, err := json.Marshal(vRef); err != nil {
			return fmt.Errorf("cannot marshal this value: %v", err)
		} else {
			fmt.Println(string(bytes))
		}
	case Pretty, Markdown:
		tableBuilder(w, v)

		if currentStyle == Pretty {
			w.SetStyle(table.StyleBold)

			w.Render()
		} else {
			w.RenderMarkdown()
		}
	}

	return nil
}
