package log

import logger "gitlab.oneitfarm.com/bifrost/cilog/v2"

type oneitfarmLogger struct{}

// Interface assertions
var _ Logger = (*oneitfarmLogger)(nil)

// NewNopLogger returns a logger that doesn't do anything.
func NewOneitfarmLogger() Logger { return &oneitfarmLogger{} }

func (oneitfarmLogger) Info(msg string, keyvals ...interface{})  {
	logger.Infow(msg, keyvals...)
}
func (oneitfarmLogger) Debug(msg string, keyvals ...interface{}) {
	logger.Debugw(msg, keyvals...)
}
func (oneitfarmLogger) Error(msg string, keyvals ...interface{}) {
	logger.Errorw(msg, keyvals...)
}

func (l *oneitfarmLogger) With(keyvals ...interface{}) Logger {
	logger.With(keyvals...)
	return l
}
