package log

import (
	"os"

	"github.com/rs/zerolog"
	zerologadapter "logur.dev/adapter/zerolog"
)

func NewDefaultLogger() *LoggerAdapter {
	l := zerolog.New(os.Stderr).With().Timestamp().Logger()
	l.Level(zerolog.DebugLevel)
	logger := zerologadapter.New(l)
	return &LoggerAdapter{
		logger: logger,
	}
}
