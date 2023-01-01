package service

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

type sideEffect = string

const (
	localCopyOnly         sideEffect = "copied without overwriting"
	localMoveOnly         sideEffect = "moved without overwriting"
	localCopyAndOverwrite sideEffect = "copied and overwrote"
	localMoveAndOverwrite sideEffect = "moved and overwrote"
)

type LocalDistributionResult struct {
	localMoveResponse
	RawJson string
}

type localMoveResponse struct {
	SourceFilePath      string     `json:"source_file_path"`
	DestinationFilePath string     `json:"destination_file_path"`
	SideEffect          sideEffect `json:"side_effect"`
}

var LocalTableBuilder = func(w table.Writer, v any) {
	resp := v.(LocalDistributionResult)

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
