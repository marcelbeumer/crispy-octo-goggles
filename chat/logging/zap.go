package logging

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// mapFields maps ZapFieldAdapter to zap.Field
func mapFields(args []any) {
	for i, v := range args {
		// println(reflect.TypeOf(v).String())
		if t, ok := v.(*ZapFieldAdapter); ok {
			args[i] = zap.Field(t.f)
		}
	}
}

// ZapLoggerAdapter implements Logger for zap.
// It seems somewhat complicated and verbose to simply wrap a logger
// but I don't see how to to do it better in Go
type ZapLoggerAdapter struct {
	logger *zap.SugaredLogger
}

func (l *ZapLoggerAdapter) Debug(args ...any) {
	mapFields(args)
	l.logger.Debug(args...)
}

func (l *ZapLoggerAdapter) Debugw(msg string, keysAndValues ...any) {
	mapFields(keysAndValues)
	l.logger.Debugw(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Info(args ...any) {
	mapFields(args)
	l.logger.Info(args...)
}

func (l *ZapLoggerAdapter) Infow(msg string, keysAndValues ...any) {
	mapFields(keysAndValues)
	l.logger.Infow(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Warn(args ...any) {
	mapFields(args)
	l.logger.Warn(args...)
}

func (l *ZapLoggerAdapter) Warnw(msg string, keysAndValues ...any) {
	mapFields(keysAndValues)
	l.logger.Warnw(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Error(args ...any) {
	mapFields(args)
	l.logger.Error(args...)
}

func (l *ZapLoggerAdapter) Errorw(msg string, keysAndValues ...any) {
	mapFields(keysAndValues)
	l.logger.Errorw(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Fatal(args ...any) {
	mapFields(args)
	l.logger.Fatal(args...)
}

func (l *ZapLoggerAdapter) Fatalw(msg string, keysAndValues ...any) {
	mapFields(keysAndValues)
	l.logger.Fatalw(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Named(s string) Logger {
	named := l.logger.Named(s)
	return &ZapLoggerAdapter{logger: named}
}

func (l *ZapLoggerAdapter) With(args ...any) Logger {
	with := l.logger.With(args...)
	return &ZapLoggerAdapter{logger: with}
}

func NewZapLogger(
	out io.Writer,
	verbose bool,
	veryVerbose bool,
) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	writer := zapcore.Lock(zapcore.AddSync(out))
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
	return zap.New(core)
}

func NewZapLoggerAdapter(
	zl *zap.Logger,
) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{logger: zl.Sugar()}
}

func RedirectStdLog(l *zap.Logger) {
	zap.RedirectStdLog(l)
}

// ZapFieldAdapter implements Field for zap.
// It seems somewhat complicated and verbose to simply wrap a logger
// but I don't see how to to do it better in Go
type ZapFieldAdapter struct{ f zap.Field }

// isField is a dummy method so we don't have an empty interface
func (a ZapFieldAdapter) isField() bool {
	return true
}

func Error(err error) Field {
	return &ZapFieldAdapter{f: zap.Error(err)}
}
