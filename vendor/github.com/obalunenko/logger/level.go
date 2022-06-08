package logger

import (
	"github.com/sirupsen/logrus"
)

// Level type.
type Level struct {
	logrus.Level
}

// IsDebug checks if level is in debug range.
func (l Level) IsDebug() bool {
	return l.Level >= logrus.DebugLevel
}
