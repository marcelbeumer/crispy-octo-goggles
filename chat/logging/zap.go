package logging

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLoggerAdapter struct {
	logger *zap.SugaredLogger
}

func (l *ZapLoggerAdapter) Debug(args ...any) {
	l.logger.Debug(args...)
}

func (l *ZapLoggerAdapter) Debugw(msg string, keysAndValues ...any) {
	l.logger.Debugw(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Info(args ...any) {
	l.logger.Info(args...)
}

func (l *ZapLoggerAdapter) Infow(msg string, keysAndValues ...any) {
	l.logger.Infow(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Warn(args ...any) {
	l.logger.Warn(args...)
}

func (l *ZapLoggerAdapter) Warnw(msg string, keysAndValues ...any) {
	l.logger.Warnw(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Error(args ...any) {
	l.logger.Error(args...)
}

func (l *ZapLoggerAdapter) Errorw(msg string, keysAndValues ...any) {
	l.logger.Errorw(msg, keysAndValues...)
}

func (l *ZapLoggerAdapter) Fatal(args ...any) {
	l.logger.Fatal(args...)
}

func (l *ZapLoggerAdapter) Fatalw(msg string, keysAndValues ...any) {
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
