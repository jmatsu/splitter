package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"io"
	"os"
	"strings"
	"time"
)

const (
	DefaultLogLevel = "info"
)

var Logger zerolog.Logger
var CmdStdout io.Writer
var CmdStderr io.Writer

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	writer := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	writer.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}

	Logger = zerolog.New(writer).With().Timestamp().Logger()
	SetLogLevel(DefaultLogLevel)

	CmdStdout = Writer(zerolog.InfoLevel)
	CmdStderr = Writer(zerolog.WarnLevel)
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

func Writer(level zerolog.Level) LeveledWriter {
	logger := Logger.Level(level)

	return LeveledWriter{
		l:     &logger,
		level: level,
	}
}

type LeveledWriter struct {
	l     *zerolog.Logger
	level zerolog.Level
}

func (l LeveledWriter) Write(p []byte) (n int, err error) {
	// borrowed from zerolog's code. this code chunk's copyrights belong to zerolog authors.
	// ref: https://github.com/rs/zerolog/blob/3543e9d94bc5ed088dd2d9ad1d19c7ccd0fa65f5/log.go#L435
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		p = p[0 : n-1]
	}
	l.l.WithLevel(l.level).CallerSkipFrame(1).Msg(string(p))
	return
}
