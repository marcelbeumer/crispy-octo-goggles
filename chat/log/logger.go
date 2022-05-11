package log

// https://pkg.go.dev/go.uber.org/zap#SugaredLogger
type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Named(s string) Logger
}
