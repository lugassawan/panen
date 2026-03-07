package presenter

import (
	"context"
	"log/slog"

	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/infra/applog"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogHandler handles debug mode toggling and log export requests.
type LogHandler struct {
	ctx      context.Context
	settings settings.Repository
	logDir   string
}

// Bind injects runtime dependencies into the handler.
func (h *LogHandler) Bind(ctx context.Context, s settings.Repository, logDir string) {
	h.ctx = ctx
	h.settings = s
	h.logDir = logDir
}

// IsDebugMode returns whether debug logging is enabled.
func (h *LogHandler) IsDebugMode() (bool, error) {
	val, err := h.settings.GetSetting(h.ctx, applog.DebugLoggingKey)
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

// SetDebugMode enables or disables debug logging and persists the setting.
func (h *LogHandler) SetDebugMode(enabled bool) error {
	val := "0"
	if enabled {
		val = "1"
		applog.SetLevel(slog.LevelDebug)
	} else {
		applog.SetLevel(slog.LevelInfo)
	}
	return h.settings.SetSetting(h.ctx, applog.DebugLoggingKey, val)
}

// ExportLogs prompts the user to choose a save path, then creates a zip of recent logs.
func (h *LogHandler) ExportLogs() (string, error) {
	path, err := runtime.SaveFileDialog(h.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "panen-logs.zip",
		Title:           "Export Logs",
		Filters: []runtime.FileFilter{
			{DisplayName: "Zip Archives", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := applog.ExportLogs(h.logDir, path, applog.LogRetentionDays); err != nil {
		return "", err
	}
	return path, nil
}

// GetLogStats returns aggregate statistics about log files.
func (h *LogHandler) GetLogStats() (*LogStatsResponse, error) {
	stats, err := applog.GetLogStats(h.logDir)
	if err != nil {
		return nil, err
	}
	return &LogStatsResponse{
		FileCount:  stats.FileCount,
		TotalBytes: stats.TotalBytes,
		OldestDate: stats.OldestDate,
		NewestDate: stats.NewestDate,
	}, nil
}
