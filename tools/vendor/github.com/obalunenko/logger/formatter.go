package logger

import (
	"time"

	"github.com/sirupsen/logrus" //nolint:depguard // this is the only place where logrus should be imported.
)

const (
	jsonFmt = "json"
	textFmt = "text"
)

func makeFormatter(format string) logrus.Formatter {
	var f logrus.Formatter

	switch format {
	case jsonFmt:
		f = jsonFormatter()
	case textFmt:
		f = textFormatter()
	default:
		f = textFormatter()
	}

	return f
}

func jsonFormatter() logrus.Formatter {
	f := new(logrus.JSONFormatter)
	f.TimestampFormat = time.RFC3339Nano

	f.DataKey = "metadata"

	return f
}

func textFormatter() logrus.Formatter {
	f := new(logrus.TextFormatter)

	f.ForceColors = true
	f.DisableColors = false
	f.FullTimestamp = true
	f.TimestampFormat = "02-01-2006 15:04:05"
	f.QuoteEmptyFields = true

	return f
}
