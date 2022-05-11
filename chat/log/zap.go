package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func (l *ZapLogger) Debug(args ...any) {
	l.logger.Debug(args...)
}

func (l *ZapLogger) Info(args ...any) {
	l.logger.Info(args...)
}

func (l *ZapLogger) Warn(args ...any) {
	l.logger.Warn(args...)
}

func (l *ZapLogger) Error(args ...any) {
	l.logger.Error(args...)
}

func (l *ZapLogger) Fatal(args ...any) {
	l.logger.Fatal(args...)
}

func (l *ZapLogger) Named(s string) Logger {
	named := l.logger.Named(s)
	return &ZapLogger{logger: named}
}

func NewZapLogger(out io.Writer, verbose bool, veryVerbose bool) (*ZapLogger, error) {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	writer := zapcore.Lock(os.Stderr)
	levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		switch {
		case veryVerbose:
			return lvl >= zap.DebugLevel
		case verbose:
			return lvl >= zap.InfoLevel
		default:
			return lvl >= zap.ErrorLevel
		}
	})
	core := zapcore.NewCore(encoder, writer, levelEnabler)
	zl := zap.New(core).Sugar()
	logger := &ZapLogger{logger: zl}
	return logger, nil
}
