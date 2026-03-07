package backup

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lugassawan/panen/backend/infra/applog"

	_ "modernc.org/sqlite"
)

// TryRecover checks the database at dataDir/panen.db and restores from backup if needed.
// Returns the backup filename that was restored from, or "" if no restore was needed/possible.
func TryRecover(dataDir, backupDir string) (string, error) {
	dbPath := filepath.Join(dataDir, "panen.db")

	if _, err := os.Stat(dbPath); err == nil {
		if err := quickCheck(dbPath); err == nil {
			return "", nil // DB exists and is healthy
		}
		applog.Warn("database corruption detected, attempting recovery", nil, applog.Fields{
			"dbPath": dbPath,
		})
	}

	svc := NewBackupService()
	backups, err := svc.ListBackups(backupDir)
	if err != nil {
		return "", fmt.Errorf("list backups for recovery: %w", err)
	}

	for _, b := range backups {
		src := filepath.Join(backupDir, b.Filename)
		if err := restoreAndValidate(src, dbPath); err != nil {
			applog.Warn("backup failed validation, trying next", err, applog.Fields{
				"backup": b.Filename,
			})
			continue
		}
		applog.Info("database restored from backup", applog.Fields{
			"backup": b.Filename,
		})
		return b.Filename, nil
	}

	// All backups failed or none exist — remove corrupt DB so migrate creates fresh
	_ = os.Remove(dbPath)
	return "", nil
}

// restoreAndValidate copies a backup to dbPath and runs quick_check.
func restoreAndValidate(src, dbPath string) error {
	if err := copyFile(src, dbPath); err != nil {
		return fmt.Errorf("copy backup: %w", err)
	}
	if err := quickCheck(dbPath); err != nil {
		_ = os.Remove(dbPath)
		return fmt.Errorf("restored db failed check: %w", err)
	}
	return nil
}

// quickCheck opens a temporary connection and runs PRAGMA quick_check.
func quickCheck(dbPath string) error {
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout%3d5000")
	if err != nil {
		return err
	}
	defer conn.Close()

	var result string
	if err := conn.QueryRow("PRAGMA quick_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("quick_check: %s", result)
	}
	return nil
}
