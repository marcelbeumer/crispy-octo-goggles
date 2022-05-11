package log

import (
	"os"

	"github.com/rs/zerolog"
	zerologadapter "logur.dev/adapter/zerolog"
)

func NewDefaultLogger(verbose bool, veryVerbose bool) *LoggerAdapter {
	l := zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Logger()

	switch {
	case veryVerbose:
		l = l.Level(zerolog.DebugLevel)
	case verbose:
		l = l.Level(zerolog.InfoLevel)
	default:
		l = l.Level(zerolog.ErrorLevel)
	}

	logger := zerologadapter.New(l)
	return &LoggerAdapter{
		logger: logger,
	}
}
