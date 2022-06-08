package log

// NoopLoggerAdapter implements Logger but does not do any logging
type NoopLoggerAdapter struct {
}

func (l *NoopLoggerAdapter) Debug(args ...any) {
}

func (l *NoopLoggerAdapter) Debugw(msg string, keysAndValues ...any) {
}

func (l *NoopLoggerAdapter) Info(args ...any) {
}

func (l *NoopLoggerAdapter) Infow(msg string, keysAndValues ...any) {
}

func (l *NoopLoggerAdapter) Warn(args ...any) {
}

func (l *NoopLoggerAdapter) Warnw(msg string, keysAndValues ...any) {
}

func (l *NoopLoggerAdapter) Error(args ...any) {
}

func (l *NoopLoggerAdapter) Errorw(msg string, keysAndValues ...any) {
}

func (l *NoopLoggerAdapter) Fatal(args ...any) {
}

func (l *NoopLoggerAdapter) Fatalw(msg string, keysAndValues ...any) {
}

func (l *NoopLoggerAdapter) Named(s string) Logger {
	return l
}

func (l *NoopLoggerAdapter) With(args ...any) Logger {
	return l
}
