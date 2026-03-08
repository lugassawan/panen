package backup

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/lugassawan/panen/backend/infra/applog"
	"github.com/lugassawan/panen/backend/infra/database"

	_ "modernc.org/sqlite"
)

var safeReason = regexp.MustCompile(`[^a-zA-Z0-9-]`)

// BackupInfo holds metadata about a single backup file.
type BackupInfo struct {
	Filename  string
	SizeBytes int64
	CreatedAt time.Time
}

// BackupService manages database backup operations.
type BackupService struct{}

// NewBackupService creates a new BackupService.
func NewBackupService() *BackupService {
	return &BackupService{}
}

// RunDaily creates a daily backup if one doesn't already exist for today.
func (s *BackupService) RunDaily(dbPath, backupDir string) error {
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	filename := fmt.Sprintf("panen-%s.db", time.Now().Format(time.DateOnly))
	dst := filepath.Join(backupDir, filename)

	if _, err := os.Stat(dst); err == nil {
		return nil // already exists for today
	}

	if err := checkpoint(dbPath); err != nil {
		return fmt.Errorf("checkpoint before daily backup: %w", err)
	}
	if err := copyFile(dbPath, dst); err != nil {
		return fmt.Errorf("copy daily backup: %w", err)
	}
	return nil
}

// Cleanup removes backups older than retentionDays and warns if total size exceeds 100MB.
func (s *BackupService) Cleanup(backupDir string, retentionDays int) error {
	backups, err := s.ListBackups(backupDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var totalSize int64
	for _, b := range backups {
		if b.CreatedAt.Before(cutoff) {
			_ = os.Remove(filepath.Join(backupDir, b.Filename))
		} else {
			totalSize += b.SizeBytes
		}
	}

	const warnThreshold = 100 * 1024 * 1024 // 100MB
	if totalSize > warnThreshold {
		applog.Warn("backup total size exceeds 100MB", nil, applog.Fields{
			"totalBytes": totalSize,
			"backupDir":  backupDir,
		})
	}
	return nil
}

// CreateBeforeDestructive creates a backup before a destructive operation.
func (s *BackupService) CreateBeforeDestructive(dbPath, backupDir, reason string) error {
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	sanitized := safeReason.ReplaceAllString(reason, "")
	if sanitized == "" {
		sanitized = "unknown"
	}

	base := fmt.Sprintf("panen-%s-pre-%s", time.Now().Format(time.DateOnly), sanitized)
	dst := uniquePath(backupDir, base)

	if err := checkpoint(dbPath); err != nil {
		return fmt.Errorf("checkpoint before destructive backup: %w", err)
	}
	return copyFile(dbPath, dst)
}

// CreateManualBackup creates a user-triggered backup.
func (s *BackupService) CreateManualBackup(dbPath, backupDir string) error {
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	base := fmt.Sprintf("panen-%s-manual", time.Now().Format(time.DateOnly))
	dst := uniquePath(backupDir, base)

	if err := checkpoint(dbPath); err != nil {
		return fmt.Errorf("checkpoint before manual backup: %w", err)
	}
	return copyFile(dbPath, dst)
}

// ListBackups returns all backup files sorted newest-first.
func (s *BackupService) ListBackups(backupDir string) ([]BackupInfo, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read backup dir: %w", err)
	}

	var backups []BackupInfo
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "panen-") || !strings.HasSuffix(name, ".db") {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Filename:  name,
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})
	return backups, nil
}

// uniquePath returns a path like base.db, or base-1.db, base-2.db if the file already exists.
func uniquePath(dir, base string) string {
	dst := filepath.Join(dir, base+".db")
	for i := 1; ; i++ {
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			return dst
		}
		dst = filepath.Join(dir, fmt.Sprintf("%s-%d.db", base, i))
	}
}

// checkpoint opens a temporary connection to run WAL checkpoint.
func checkpoint(dbPath string) error {
	conn, err := sql.Open(database.SQLiteDriver, dbPath+"?_pragma=busy_timeout%3d5000")
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// copyFile copies src to dst using io.Copy, syncing to disk before closing.
// On failure, the partial destination file is removed.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}
	return out.Close()
}
