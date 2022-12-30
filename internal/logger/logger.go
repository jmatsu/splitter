package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strings"
	"time"
)

var Logger zerolog.Logger

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	writer.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}

	Logger = zerolog.New(writer).With().Timestamp().Logger()
}

func SetDebugMode() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}
