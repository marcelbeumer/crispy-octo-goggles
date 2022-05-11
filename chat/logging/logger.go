package logging

import (
	"logur.dev/logur"
)

type Logger interface {
	Trace(msg string, fields ...map[string]any)
	Debug(msg string, fields ...map[string]any)
	Info(msg string, fields ...map[string]any)
	Warn(msg string, fields ...map[string]any)
	Error(msg string, fields ...map[string]any)
	WithFields(fields map[string]any) Logger
}

type LoggerAdapter struct {
	logger logur.Logger
}

func (n *LoggerAdapter) Trace(msg string, fields ...map[string]any) {
	n.logger.Trace(msg, fields...)
}

func (n *LoggerAdapter) Debug(msg string, fields ...map[string]any) {
	n.logger.Debug(msg, fields...)
}

func (n *LoggerAdapter) Info(msg string, fields ...map[string]any) {
	n.logger.Info(msg, fields...)
}

func (n *LoggerAdapter) Warn(msg string, fields ...map[string]any) {
	n.logger.Warn(msg, fields...)
}

func (n *LoggerAdapter) Error(msg string, fields ...map[string]any) {
	n.logger.Error(msg, fields...)
}

func (n *LoggerAdapter) WithFields(fields map[string]any) Logger {
	return &LoggerAdapter{logur.WithFields(n.logger, fields)}
}
