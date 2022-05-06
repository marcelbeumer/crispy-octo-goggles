package log

import (
	"github.com/sirupsen/logrus"
	logrusadapter "logur.dev/adapter/logrus"
)

func NewDefaultLogger() *LoggerAdapter {
	lr := logrus.New()
	lr.SetLevel(logrus.DebugLevel)
	return &LoggerAdapter{
		logger: logrusadapter.New(lr),
	}
}
