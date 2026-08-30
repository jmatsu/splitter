package internal

import "time"

// timestampLayout must be kept in sync with the layout .goreleaser.yaml stamps Timestamp with.
const timestampLayout = "2006-01-02 15:04:05"

var (
	Version    = "undefined"
	Commit     = "undefined"
	Timestamp  = "undefined"
	CompiledAt time.Time
)

func init() {
	var err error

	CompiledAt, err = time.Parse(timestampLayout, Timestamp)

	if err != nil {
		CompiledAt = time.Now()
	}
}
