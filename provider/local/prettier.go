package local

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

type sideEffect = string

const (
	copyOnly         sideEffect = "copied without overwriting"
	moveOnly         sideEffect = "moved without overwriting"
	copyAndOverwrite sideEffect = "copied and overwrote"
	moveAndOverwrite sideEffect = "moved and overwrote"
)

type DistributionResult struct {
	moveResponse
	RawJson string
}

type moveResponse struct {
	SourceFilePath      string     `json:"source_file_path"`
	DestinationFilePath string     `json:"destination_file_path"`
	SideEffect          sideEffect `json:"side_effect"`
}

var TableBuilder = func(w table.Writer, v any) {
	resp := v.(DistributionResult)

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
