package task

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/internal/config"
	"testing"
)

func Test_Formatter_Format(t *testing.T) {
	cases := map[string]struct {
		formatter *Formatter
	}{
		"with table and style": {
			formatter: NewFormatter().withTableBuilder(func(writer table.Writer, r any) {
				// no-op
			}).withStyle(config.PrettyFormat),
		},
		"with table": {
			formatter: NewFormatter().withTableBuilder(func(writer table.Writer, r any) {
				// no-op
			}),
		},
		"zero": {
			formatter: NewFormatter(),
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			if err := c.formatter.Format(&testDistributionResult{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type testDistributionResult struct {
}

func (r *testDistributionResult) ValueResponse() any {
	return struct{}{}
}

func (r *testDistributionResult) RawJsonResponse() string {
	return "ok"
}

func (f *Formatter) withTableBuilder(b TableBuilder) *Formatter {
	f.TableBuilder = b
	return f
}

func (f *Formatter) withStyle(s config.FormatStyle) *Formatter {
	f.style = s
	return f
}
