package test

import (
	"os"

	"github.com/marcelbeumer/go-playground/gochat/internal/log"
)

func NewTestLogger(silent bool) log.Logger {
	if silent {
		return &log.NoopLoggerAdapter{}
	}
	zl := log.NewZapLogger(os.Stderr, true, true)
	return log.NewZapLoggerAdapter(zl)
}
