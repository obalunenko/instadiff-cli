package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

var logInstance *slog.Logger

func init() {
	w := os.Stderr
	lvl := LevelInfo

	h := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   false,
		Level:       lvl,
		ReplaceAttr: replaceLevelNames,
	})

	l := slog.New(h)

	logInstance = l
}

// Params holds logger specific params.
type Params struct {
	// Writer is a writer to write logs to. By default, it's os.Stderr.
	Writer io.WriteCloser
	// Level is one of "debug", "info", "warn", "error", "fatal".
	Level string
	// Format is one of "json" or "text".
	Format string
	// WithSource enables source code info in logs.
	WithSource bool
}

// Init initiates logger and add format options.
// Should be called only once, on start of app.
func Init(ctx context.Context, p Params) Logger {
	if p.Writer == nil {
		p.Writer = os.Stderr
	}

	return makeLogInstance(ctx, p)
}

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// Logger serves as an adapter interface for logger libraries
// so that we not depend on any of them directly.
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)

	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
}

func makeLogInstance(ctx context.Context, p Params) Logger {
	level, err := ParseLevel(p.Level)
	if err != nil {
		WithError(ctx, err).Error("unable to parse log level")
		level = LevelInfo
	}

	formatter := makeFormatter(p.Writer, p.Format, level, p.WithSource, replaceLevelNames)

	logInstance = slog.New(formatter)

	levels := make([]string, 0, len(AllLevels))

	for _, l := range AllLevels {
		if logInstance.Enabled(ctx, l) {
			levels = append(levels, l.String())
		}
	}

	WithField(ctx, "levels", strings.Join(levels, " ")).Debug("logging enabled")

	return newSlogWrapper(logInstance)
}

var _ Logger = (*slogWrapper)(nil)

func newSlogWrapper(entry *slog.Logger) *slogWrapper {
	return &slogWrapper{
		le: entry,
	}
}

type slogWrapper struct {
	le *slog.Logger
}

func (s slogWrapper) Debug(msg string) {
	s.le.Debug(msg)
}

func (s slogWrapper) Info(msg string) {
	s.le.Info(msg)
}

func (s slogWrapper) Warn(msg string) {
	s.le.Warn(msg)
}

func (s slogWrapper) Error(msg string) {
	s.le.Error(msg)
}

func (s slogWrapper) Fatal(msg string) {
	s.le.Log(context.Background(), LevelFatal, msg)

	os.Exit(1)
}

func (s slogWrapper) WithError(err error) Logger {
	s.le.With()
	return newSlogWrapper(s.le.With("error", err))
}

func (s slogWrapper) WithField(key string, value interface{}) Logger {
	return newSlogWrapper(s.le.With(key, value))
}

func (s slogWrapper) WithFields(fields Fields) Logger {
	attrs := make([]any, 0, len(fields))

	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	return newSlogWrapper(s.le.With(attrs...))
}
