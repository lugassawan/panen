package applog

import (
	"log/slog"
	"runtime"
	"strings"
)

// Fields is a map of structured log attributes.
type Fields map[string]any

// Info logs an informational message with automatic caller context.
func Info(msg string, fields Fields) {
	slog.Info(msg, toAttrs(fields)...)
}

// Warn logs a warning with automatic caller and error context.
func Warn(msg string, err error, fields Fields) {
	attrs := toAttrs(fields)
	attrs = append(attrs, "err", err)
	slog.Warn(msg, attrs...)
}

// Error logs an error with automatic caller and error context.
func Error(msg string, err error, fields Fields) {
	attrs := toAttrs(fields)
	attrs = append(attrs, "err", err)
	slog.Error(msg, attrs...)
}

func toAttrs(fields Fields) []any {
	attrs := callerAttrs()
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return attrs
}

func callerAttrs() []any {
	pc, _, _, ok := runtime.Caller(3)
	if !ok {
		return nil
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nil
	}
	return []any{"caller", shortCaller(fn.Name())}
}

func shortCaller(full string) string {
	const prefix = "github.com/lugassawan/panen/backend/"
	after, found := strings.CutPrefix(full, prefix)
	if !found {
		return full
	}
	return after
}
