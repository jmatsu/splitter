package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"strings"
	"time"
)

const (
	DefaultLogLevel = "warn"
)

var Logger zerolog.Logger

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	writer := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	writer.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}

	Logger = zerolog.New(writer).With().Timestamp().Logger()
}

func SetLogLevel(level string) {
	var logLevel zerolog.Level

	switch strings.ToLower(level) {
	case "trace":
		logLevel = zerolog.TraceLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.WarnLevel
	}

	zerolog.SetGlobalLevel(logLevel)
}
