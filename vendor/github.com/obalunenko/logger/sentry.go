package logger

import (
	"context"
	"fmt"
	"strings"

	"github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"
)

// SentryParams holds sentry specific params.
type SentryParams struct {
	Enabled      bool
	DSN          string
	TraceEnabled bool
	TraceLevel   string
	Tags         map[string]string
}

func setupSentry(ctx context.Context, p SentryParams) error {
	hook, err := logrus_sentry.NewAsyncWithTagsSentryHook(p.DSN, p.Tags, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	if err != nil {
		return fmt.Errorf("setup async hook: %w", err)
	}

	if p.TraceEnabled {
		traceLevel, err := logrus.ParseLevel(p.TraceLevel)
		if err != nil {
			traceLevel = logrus.PanicLevel
		}

		hook.StacktraceConfiguration.Enable = true
		hook.StacktraceConfiguration.Level = traceLevel
	}

	levels := make([]string, 0, len(hook.Levels()))

	for _, l := range hook.Levels() {
		levels = append(levels, l.String())
	}

	WithField(ctx, "levels", strings.Join(levels, " ")).Info("sentry enabled")

	logInstance.AddHook(hook)

	return nil
}
