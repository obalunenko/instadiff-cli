package logger

import (
	"fmt"
	"log/slog"
	"strings"
)

// Level type.
type Level = slog.Level

const (
	// LevelDebug is a debug level.
	LevelDebug = slog.LevelDebug
	// LevelInfo is an info level.
	LevelInfo = slog.LevelInfo
	// LevelWarn is a warn level.
	LevelWarn = slog.LevelWarn
	// LevelError is an error level.
	LevelError = slog.LevelError
	// LevelFatal is a fatal level.
	LevelFatal = slog.Level(12)
)

// AllLevels is a slice of all levels.
var AllLevels = []Level{
	LevelDebug,
	LevelInfo,
	LevelWarn,
	LevelError,
	LevelFatal,
}

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

var levelValues = map[string]Level{
	"DEBUG": LevelDebug,
	"INFO":  LevelInfo,
	"WARN":  LevelWarn,
	"ERROR": LevelError,
	"FATAL": LevelFatal,
}

func replaceLevelNames(groups []string, a slog.Attr) slog.Attr {
	// Customize the name of the level key and the output string, including
	// custom level values.
	if a.Key == slog.LevelKey {
		level, ok := a.Value.Any().(slog.Level)
		if !ok {
			return a
		}

		name, ok := levelNames[level]
		if !ok {
			name = "UNKNOWN"
		}

		a.Value = slog.StringValue(name)
	}

	return a
}

// ParseLevel takes a string level and returns the Logrus log level constant.
func ParseLevel(lvl string) (Level, error) {
	level, ok := levelValues[strings.ToUpper(lvl)]
	if !ok {
		return LevelInfo, fmt.Errorf("not a valid log Level: %q", lvl)
	}

	return level, nil
}
