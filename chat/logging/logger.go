package logging

// https://pkg.go.dev/go.uber.org/zap#SugaredLogger
type Logger interface {
	Debug(args ...any)
	Debugw(msg string, keysAndValues ...any)
	Info(args ...any)
	Infow(msg string, keysAndValues ...any)
	Warn(args ...any)
	Warnw(msg string, keysAndValues ...any)
	Error(args ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatal(args ...any)
	Fatalw(msg string, keysAndValues ...any)
	Named(s string) Logger
	With(args ...any) Logger
}
