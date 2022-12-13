package logger

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus" //nolint:depguard // this is the only place where logrus should be imported.
)

var (
	logInstance *logrus.Logger
)

func init() {
	logInstance = logrus.New()
}

// Params holds logger specific params.
type Params struct {
	Writer       io.WriteCloser
	Level        string
	Format       string
	SentryParams SentryParams
}

// Init initiates logger and add format options.
// Should be called only once, on start of app.
func Init(ctx context.Context, p Params) {
	if p.Writer == nil {
		p.Writer = os.Stderr
	}

	makeLogInstance(ctx, p)
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

	Writer() io.WriteCloser
	LogLevel() Level
}

type logrusWrapper struct {
	le *logrus.Entry
}

func (l logrusWrapper) LogLevel() Level {
	return Level{Level: logInstance.GetLevel()}
}

func (l logrusWrapper) Debug(msg string) {
	l.le.Debug(msg)
}

func (l logrusWrapper) Info(msg string) {
	l.le.Info(msg)
}

func (l logrusWrapper) Warn(msg string) {
	l.le.Warn(msg)
}

func (l logrusWrapper) Error(msg string) {
	l.le.Error(msg)
}

func (l logrusWrapper) Fatal(msg string) {
	l.le.Fatal(msg)
}

func (l logrusWrapper) WithError(err error) Logger {
	return newLogrusWrapper(l.le.WithError(err))
}

func (l logrusWrapper) WithField(key string, value interface{}) Logger {
	return newLogrusWrapper(l.le.WithField(key, value))
}

func (l logrusWrapper) WithFields(fields Fields) Logger {
	return newLogrusWrapper(l.le.WithFields(logrus.Fields(fields)))
}

func (l logrusWrapper) Writer() io.WriteCloser {
	return l.le.Writer()
}

func newLogrusWrapper(entry *logrus.Entry) *logrusWrapper {
	return &logrusWrapper{
		le: entry,
	}
}

func makeLogInstance(ctx context.Context, p Params) {
	level, err := logrus.ParseLevel(p.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logInstance.SetLevel(level)

	logInstance.SetFormatter(makeFormatter(p.Format))

	var out []io.Writer

	out = append(out, os.Stdout)

	logInstance.SetOutput(io.MultiWriter(out...))

	if p.SentryParams.Enabled {
		if err := setupSentry(ctx, p.SentryParams); err != nil {
			WithError(ctx, err).Error("unable to setup sentry")
		}
	}

	levels := make([]string, 0, len(logrus.AllLevels))

	for _, l := range logrus.AllLevels {
		if logInstance.IsLevelEnabled(l) {
			levels = append(levels, l.String())
		}
	}

	WithField(ctx, "levels", strings.Join(levels, " ")).Debug("logging enabled")
}
