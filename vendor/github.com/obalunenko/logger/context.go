package logger

import (
	"context"
)

type logCtxKey struct{}

// ContextWithLogger adds Logger to context and returns new context.
func ContextWithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, logCtxKey{}, l)
}

// FromContext extracts Logger from context. If no instance found - returns Logger from default logInstance.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return newSlogWrapper(logInstance)
	}

	if l, ok := ctx.Value(logCtxKey{}).(Logger); ok && l != nil {
		return l
	}

	return newSlogWrapper(logInstance)
}
