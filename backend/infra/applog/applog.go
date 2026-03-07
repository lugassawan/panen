package applog

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Fields is a map of structured log attributes.
type Fields map[string]any

// LogStats holds aggregate statistics about log files.
type LogStats struct {
	FileCount  int
	TotalBytes int64
	OldestDate string
	NewestDate string
}

const (
	logPrefix  = "panen-"
	logSuffix  = ".log"
	dateLayout = "2006-01-02"
)

var (
	level   slog.LevelVar // atomic, default Info
	logFile *os.File
)

// Init sets up file-based JSON structured logging.
// Logs are written to logDir/panen-YYYY-MM-DD.log and also to stderr.
func Init(logDir string) error {
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	today := time.Now().Format(dateLayout)
	name := filepath.Join(logDir, logPrefix+today+logSuffix)

	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	logFile = f

	handler := slog.NewJSONHandler(
		io.MultiWriter(logFile, os.Stderr),
		&slog.HandlerOptions{Level: &level},
	)
	slog.SetDefault(slog.New(handler))
	return nil
}

// Shutdown closes the log file.
func Shutdown() {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// Debug logs a debug-level message with automatic caller context.
func Debug(msg string, fields Fields) {
	slog.Debug(msg, toAttrs(fields)...)
}

// SetLevel changes the minimum log level at runtime (atomic).
func SetLevel(l slog.Level) {
	level.Set(l)
}

// Level returns the current minimum log level.
func Level() slog.Level {
	return level.Level()
}

// RotateLogs removes log files in logDir older than retentionDays.
func RotateLogs(logDir string, retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return walkLogFiles(logDir, func(path, date string) error {
		t, parseErr := time.Parse(dateLayout, date)
		if parseErr != nil {
			return nil //nolint:nilerr // skip files with unparseable dates
		}
		if t.Before(cutoff) {
			return os.Remove(path)
		}
		return nil
	})
}

// ExportLogs creates a zip at exportPath containing log files from the last days.
// The active log file may be partially captured if writes occur during export.
func ExportLogs(logDir, exportPath string, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)

	out, err := os.Create(exportPath)
	if err != nil {
		return fmt.Errorf("create export file: %w", err)
	}

	w := zip.NewWriter(out)

	walkErr := walkLogFiles(logDir, func(path, date string) error {
		t, parseErr := time.Parse(dateLayout, date)
		if parseErr != nil {
			return nil //nolint:nilerr // skip files with unparseable dates
		}
		if t.Before(cutoff) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f, err := w.Create(filepath.Base(path))
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	})

	if closeErr := w.Close(); closeErr != nil && walkErr == nil {
		walkErr = closeErr
	}
	if closeErr := out.Close(); closeErr != nil && walkErr == nil {
		walkErr = closeErr
	}
	if walkErr != nil {
		_ = os.Remove(exportPath)
		return walkErr
	}
	return nil
}

// GetLogStats returns aggregate statistics for log files in logDir.
func GetLogStats(logDir string) (LogStats, error) {
	var stats LogStats
	var dates []string

	err := walkLogFiles(logDir, func(path, date string) error {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		stats.FileCount++
		stats.TotalBytes += info.Size()
		dates = append(dates, date)
		return nil
	})
	if err != nil {
		return stats, err
	}

	if len(dates) > 0 {
		sort.Strings(dates)
		stats.OldestDate = dates[0]
		stats.NewestDate = dates[len(dates)-1]
	}

	return stats, nil
}

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

// walkLogFiles iterates over panen-*.log files in logDir, extracting the date portion.
func walkLogFiles(logDir string, fn func(path, date string) error) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, logPrefix) || !strings.HasSuffix(name, logSuffix) {
			continue
		}
		date := strings.TrimPrefix(name, logPrefix)
		date = strings.TrimSuffix(date, logSuffix)
		if err := fn(filepath.Join(logDir, name), date); err != nil {
			return err
		}
	}
	return nil
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
