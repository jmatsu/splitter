package task

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jmatsu/splitter/service"
	"testing"
)

func Test_testFlightTableBuilder(t *testing.T) {
	cases := map[string]struct {
		result service.TestFlightDeployResult
	}{
		"zero 1": {
			result: service.TestFlightDeployResult{},
		},
		"regular": {
			result: service.TestFlightDeployResult{
				TestFlightUploadAppResponse: service.TestFlightUploadAppResponse{},
			},
		},
	}

	for name, c := range cases {
		name, c := name, c

		t.Run(name, func(t *testing.T) {
			w := table.NewWriter()

			// no panic is ok
			testFlightTableBuilder(w, c.result)
		})
	}
}
