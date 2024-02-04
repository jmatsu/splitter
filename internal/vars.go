package internal

import "time"

var (
	Version    = "undefined"
	Commit     = "undefined"
	Timestamp  = "undefined"
	CompiledAt time.Time
)

func init() {
	var err error

	CompiledAt, err = time.Parse("", Timestamp)

	if err != nil {
		CompiledAt = time.Now()
	}
}
