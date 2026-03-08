package presenter

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lugassawan/panen/backend/infra/backup"
)

// BackupHandler handles backup management requests.
type BackupHandler struct {
	ctx       context.Context
	backup    *backup.BackupService
	dbPath    string
	backupDir string
}

// Bind injects runtime dependencies into the handler.
func (h *BackupHandler) Bind(ctx context.Context, backupSvc *backup.BackupService, dbPath, backupDir string) {
	h.ctx = ctx
	h.backup = backupSvc
	h.dbPath = dbPath
	h.backupDir = backupDir
}

// CreateManualBackup creates a user-triggered backup.
func (h *BackupHandler) CreateManualBackup() error {
	if err := h.backup.CreateManualBackup(h.dbPath, h.backupDir); err != nil {
		return fmt.Errorf("create manual backup: %w", err)
	}
	return nil
}

// ListBackups returns all backups as DTOs.
func (h *BackupHandler) ListBackups() ([]BackupInfoResponse, error) {
	backups, err := h.backup.ListBackups(h.backupDir)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	result := make([]BackupInfoResponse, len(backups))
	for i, b := range backups {
		result[i] = BackupInfoResponse{
			Filename:  b.Filename,
			SizeBytes: b.SizeBytes,
			CreatedAt: b.CreatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

// GetBackupStatus returns summary info about backups.
func (h *BackupHandler) GetBackupStatus() (*BackupStatusResponse, error) {
	backups, err := h.backup.ListBackups(h.backupDir)
	if err != nil {
		return nil, fmt.Errorf("get backup status: %w", err)
	}

	resp := &BackupStatusResponse{
		BackupCount: len(backups),
	}

	if len(backups) > 0 {
		resp.LastBackupDate = backups[0].CreatedAt.Format(time.RFC3339)
	}

	for _, b := range backups {
		resp.TotalSizeBytes += b.SizeBytes
	}

	if info, err := os.Stat(h.dbPath); err == nil {
		resp.DbSizeBytes = info.Size()
	}

	return resp, nil
}
