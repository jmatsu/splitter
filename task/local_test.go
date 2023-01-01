package task

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/service"
	"testing"
)

func Test_localTableBuilder(t *testing.T) {
	cases := map[string]struct {
		result service.LocalDeployResult
	}{
		"zero 1": {
			result: service.LocalDeployResult{},
		},
		"regular": {
			result: service.LocalDeployResult{
				LocalMoveResponse: service.LocalMoveResponse{
					SourceFilePath:      "path/to/src",
					DestinationFilePath: "path/to/dest",
					SideEffect:          "side effect",
				},
			},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			w := table.NewWriter()

			// no panic is ok
			localTableBuilder(w, c.result)
		})
	}
}
