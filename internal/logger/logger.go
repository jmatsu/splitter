package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"strings"
)

const (
	DefaultLogLevel = "warn"
)

var Logger zerolog.Logger

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	writer := zerolog.ConsoleWriter{Out: os.Stderr}
	writer.FormatTimestamp = func(i interface{}) string {
		return ""
	}

	Logger = zerolog.New(writer)
	SetLogLevel(DefaultLogLevel)
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

	Logger = Logger.Level(logLevel)
}
