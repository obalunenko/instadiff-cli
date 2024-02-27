package logger

import (
	"io"
	"log/slog"
)

const (
	jsonFmt = "json"
	textFmt = "text"
)

func makeFormatter(w io.Writer, format string, lvl Level, source bool, fn ...replaceFn) slog.Handler {
	var builder handlerBuildFn
	switch format {
	case jsonFmt:
		builder = jsonFormatter()
	case textFmt:
		builder = textFormatter()
	default:
		builder = textFormatter()
	}

	return buildFormatter(builder, w, lvl, source, fn)
}

type replaceFn func(groups []string, a slog.Attr) slog.Attr

func buildFormatter(builder handlerBuildFn, w io.Writer, level Level, withSource bool, replaceFns []replaceFn) slog.Handler {
	replaceAttrs := func(replacenFns []replaceFn) replaceFn {
		return func(groups []string, a slog.Attr) slog.Attr {
			for _, fn := range replacenFns {
				a = fn(groups, a)
			}

			return a
		}
	}

	opts := slog.HandlerOptions{
		AddSource:   withSource,
		Level:       level,
		ReplaceAttr: replaceAttrs(replaceFns),
	}

	handler := builder(w, &opts)

	return handler
}

func jsonFormatter() handlerBuildFn {
	return func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewJSONHandler(w, opts)
	}
}

func textFormatter() handlerBuildFn {
	return func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewTextHandler(w, opts)
	}
}

type handlerBuildFn func(w io.Writer, opts *slog.HandlerOptions) slog.Handler
