package logger

import (
	"context"
)

// Debug prints debug-level log message. Usually very verbose and shown only when debug level is enabled.
func Debug(ctx context.Context, msg string) {
	FromContext(ctx).Debug(msg)
}

// Info prints info-level log message.
func Info(ctx context.Context, msg string) {
	FromContext(ctx).Info(msg)
}

// Warn prints warn-level log message.
func Warn(ctx context.Context, msg string) {
	FromContext(ctx).Warn(msg)
}

// Error prints error-level log message.
func Error(ctx context.Context, msg string) {
	FromContext(ctx).Error(msg)
}

// Fatal prints fatal-level log message and exit the program with code 1.
func Fatal(ctx context.Context, msg string) {
	FromContext(ctx).Fatal(msg)
}

// WithError adds an error as single field to the Logger Entry.
func WithError(ctx context.Context, err error) Logger {
	return FromContext(ctx).WithError(err)
}

// WithField adds s single field to the Logger Entry.
func WithField(ctx context.Context, key string, value interface{}) Logger {
	return FromContext(ctx).WithField(key, value)
}

// WithFields adds fields key-value to the Logger Entry.
func WithFields(ctx context.Context, fields map[string]interface{}) Logger {
	return FromContext(ctx).WithFields(fields)
}
