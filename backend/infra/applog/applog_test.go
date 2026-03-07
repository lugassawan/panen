package applog

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTestLog(t *testing.T, dir, date, content string) {
	t.Helper()
	path := filepath.Join(dir, "panen-"+date+".log")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test log %s: %v", path, err)
	}
}

func TestShortCaller(t *testing.T) {
	tests := []struct {
		name string
		full string
		want string
	}{
		{
			name: "trims module prefix",
			full: "github.com/lugassawan/panen/backend/usecase.(*PortfolioService).syncHoldingPeak",
			want: "usecase.(*PortfolioService).syncHoldingPeak",
		},
		{
			name: "trims infra prefix",
			full: "github.com/lugassawan/panen/backend/infra/db.(*SQLite).Query",
			want: "infra/db.(*SQLite).Query",
		},
		{
			name: "returns full if no prefix match",
			full: "fmt.Println",
			want: "fmt.Println",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortCaller(tt.full)
			if got != tt.want {
				t.Errorf("shortCaller(%q) = %q, want %q", tt.full, got, tt.want)
			}
		})
	}
}

func setTestLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	original := slog.Default()
	t.Cleanup(func() { slog.SetDefault(original) })
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	return &buf
}

func TestInfo(t *testing.T) {
	buf := setTestLogger(t)

	Info("test info", Fields{"key": "val"})

	out := buf.String()
	if !strings.Contains(out, "test info") {
		t.Errorf("expected log to contain message, got %q", out)
	}
	if !strings.Contains(out, "caller=infra/applog.TestInfo") {
		t.Errorf("expected caller to point to TestInfo, got %q", out)
	}
	if strings.Contains(out, "err=") {
		t.Errorf("info log should not contain err key, got %q", out)
	}
}

func TestWarn(t *testing.T) {
	buf := setTestLogger(t)

	Warn("test warn", errors.New("boom"), Fields{"key": "val"})

	out := buf.String()
	if !strings.Contains(out, "test warn") {
		t.Errorf("expected log to contain message, got %q", out)
	}
	if !strings.Contains(out, "caller=infra/applog.TestWarn") {
		t.Errorf("expected caller to point to TestWarn, got %q", out)
	}
	if !strings.Contains(out, "err=boom") {
		t.Errorf("expected log to contain err=boom, got %q", out)
	}
}

func TestError(t *testing.T) {
	buf := setTestLogger(t)

	Error("test error", errors.New("fail"), nil)

	out := buf.String()
	if !strings.Contains(out, "test error") {
		t.Errorf("expected log to contain message, got %q", out)
	}
	if !strings.Contains(out, "caller=infra/applog.TestError") {
		t.Errorf("expected caller to point to TestError, got %q", out)
	}
	if !strings.Contains(out, "err=fail") {
		t.Errorf("expected log to contain err=fail, got %q", out)
	}
}

func TestInit(t *testing.T) {
	dir := t.TempDir()
	original := slog.Default()
	t.Cleanup(func() {
		Shutdown()
		slog.SetDefault(original)
	})

	if err := Init(dir); err != nil {
		t.Fatalf("Init(%q) error: %v", dir, err)
	}

	Info("hello from init", Fields{"test": true})

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(dir, "panen-"+today+".log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("log line not valid JSON: %v\nraw: %s", err, data)
	}
	if msg, ok := entry["msg"].(string); !ok || msg != "hello from init" {
		t.Errorf("expected msg 'hello from init', got %v", entry["msg"])
	}
}

func TestDebugNotWrittenAtInfoLevel(t *testing.T) {
	dir := t.TempDir()
	original := slog.Default()
	t.Cleanup(func() {
		Shutdown()
		slog.SetDefault(original)
	})

	if err := Init(dir); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	Debug("secret debug msg", Fields{"x": 1})

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(dir, "panen-"+today+".log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if strings.Contains(string(data), "secret debug msg") {
		t.Error("debug message should not appear at Info level")
	}
}

func TestSetLevelEnablesDebug(t *testing.T) {
	dir := t.TempDir()
	original := slog.Default()
	t.Cleanup(func() {
		Shutdown()
		SetLevel(slog.LevelInfo) // reset
		slog.SetDefault(original)
	})

	if err := Init(dir); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	SetLevel(slog.LevelDebug)

	if got := Level(); got != slog.LevelDebug {
		t.Errorf("Level() = %v, want %v", got, slog.LevelDebug)
	}

	Debug("visible debug", Fields{"y": 2})

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(dir, "panen-"+today+".log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "visible debug") {
		t.Error("debug message should appear after SetLevel(Debug)")
	}
}

func TestRotateLogs(t *testing.T) {
	dir := t.TempDir()

	old := time.Now().AddDate(0, 0, -20).Format("2006-01-02")
	recent := time.Now().Format("2006-01-02")
	writeTestLog(t, dir, old, "old")
	writeTestLog(t, dir, recent, "recent")

	if err := RotateLogs(dir, 14); err != nil {
		t.Fatalf("RotateLogs error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "panen-"+old+".log")); !os.IsNotExist(err) {
		t.Error("old log file should have been removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "panen-"+recent+".log")); err != nil {
		t.Error("recent log file should have been preserved")
	}
}

func TestRotateLogsPreservesRecentFiles(t *testing.T) {
	dir := t.TempDir()

	for i := range 5 {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		writeTestLog(t, dir, date, "data")
	}

	if err := RotateLogs(dir, 14); err != nil {
		t.Fatalf("RotateLogs error: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 5 {
		t.Errorf("expected 5 log files preserved, got %d", len(entries))
	}
}

func TestExportLogs(t *testing.T) {
	dir := t.TempDir()
	exportDir := t.TempDir()

	for i := range 3 {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		writeTestLog(t, dir, date, "log entry for "+date)
	}

	exportPath := filepath.Join(exportDir, "logs.zip")
	if err := ExportLogs(dir, exportPath, 7); err != nil {
		t.Fatalf("ExportLogs error: %v", err)
	}

	r, err := zip.OpenReader(exportPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 3 {
		t.Errorf("expected 3 files in zip, got %d", len(r.File))
	}
}

func TestExportLogsSkipsOldFiles(t *testing.T) {
	dir := t.TempDir()
	exportDir := t.TempDir()

	recent := time.Now().Format("2006-01-02")
	old := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	writeTestLog(t, dir, recent, "recent")
	writeTestLog(t, dir, old, "old")

	exportPath := filepath.Join(exportDir, "logs.zip")
	if err := ExportLogs(dir, exportPath, 7); err != nil {
		t.Fatalf("ExportLogs error: %v", err)
	}

	r, err := zip.OpenReader(exportPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Errorf("expected 1 file in zip (only recent), got %d", len(r.File))
	}
}

func TestGetLogStats(t *testing.T) {
	dir := t.TempDir()

	dates := []string{
		time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
		time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		time.Now().Format("2006-01-02"),
	}
	var totalSize int64
	for _, d := range dates {
		content := "log data for " + d
		writeTestLog(t, dir, d, content)
		totalSize += int64(len(content))
	}

	stats, err := GetLogStats(dir)
	if err != nil {
		t.Fatalf("GetLogStats error: %v", err)
	}

	if stats.FileCount != 3 {
		t.Errorf("FileCount = %d, want 3", stats.FileCount)
	}
	if stats.TotalBytes != totalSize {
		t.Errorf("TotalBytes = %d, want %d", stats.TotalBytes, totalSize)
	}
	if stats.OldestDate != dates[0] {
		t.Errorf("OldestDate = %q, want %q", stats.OldestDate, dates[0])
	}
	if stats.NewestDate != dates[2] {
		t.Errorf("NewestDate = %q, want %q", stats.NewestDate, dates[2])
	}
}

func TestGetLogStatsEmptyDir(t *testing.T) {
	dir := t.TempDir()

	stats, err := GetLogStats(dir)
	if err != nil {
		t.Fatalf("GetLogStats error: %v", err)
	}
	if stats.FileCount != 0 {
		t.Errorf("FileCount = %d, want 0", stats.FileCount)
	}
	if stats.OldestDate != "" || stats.NewestDate != "" {
		t.Errorf("expected empty dates, got oldest=%q newest=%q", stats.OldestDate, stats.NewestDate)
	}
}
