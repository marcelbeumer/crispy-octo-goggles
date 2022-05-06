package log

import (
	"os"

	"github.com/rs/zerolog"
	zerologadapter "logur.dev/adapter/zerolog"
)

func NewDefaultLogger() *LoggerAdapter {
	l := zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)
	logger := zerologadapter.New(l)
	return &LoggerAdapter{
		logger: logger,
	}
}
