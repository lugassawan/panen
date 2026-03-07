package backup

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// createTestDB creates a minimal SQLite database file for testing.
func createTestDB(t *testing.T, dir string) string {
	t.Helper()
	dbPath := filepath.Join(dir, "panen.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	_, err = conn.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("create test table: %v", err)
	}
	conn.Close()
	return dbPath
}

func TestRunDaily(t *testing.T) {
	svc := NewBackupService()

	t.Run("creates daily backup", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := createTestDB(t, dir)
		backupDir := filepath.Join(dir, "backups")

		if err := svc.RunDaily(dbPath, backupDir); err != nil {
			t.Fatalf("RunDaily() error = %v", err)
		}

		expected := filepath.Join(backupDir, "panen-"+time.Now().Format(time.DateOnly)+".db")
		if _, err := os.Stat(expected); err != nil {
			t.Errorf("expected backup file %s not found", expected)
		}
	})

	t.Run("skips if today backup exists", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := createTestDB(t, dir)
		backupDir := filepath.Join(dir, "backups")
		if err := os.MkdirAll(backupDir, 0o750); err != nil {
			t.Fatalf("create backup dir: %v", err)
		}

		// Pre-create today's backup
		todayFile := filepath.Join(backupDir, "panen-"+time.Now().Format(time.DateOnly)+".db")
		if err := os.WriteFile(todayFile, []byte("existing"), 0o644); err != nil {
			t.Fatalf("create existing backup: %v", err)
		}

		if err := svc.RunDaily(dbPath, backupDir); err != nil {
			t.Fatalf("RunDaily() error = %v", err)
		}

		// Verify the existing file was NOT overwritten
		data, _ := os.ReadFile(todayFile)
		if string(data) != "existing" {
			t.Error("existing backup was overwritten")
		}
	})

	t.Run("creates backup dir if missing", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := createTestDB(t, dir)
		backupDir := filepath.Join(dir, "nested", "backups")

		if err := svc.RunDaily(dbPath, backupDir); err != nil {
			t.Fatalf("RunDaily() error = %v", err)
		}

		if _, err := os.Stat(backupDir); err != nil {
			t.Error("backup dir was not created")
		}
	})
}

func TestCleanup(t *testing.T) {
	svc := NewBackupService()

	t.Run("removes old backups", func(t *testing.T) {
		dir := t.TempDir()

		// Create an old backup with ModTime 10 days ago
		oldFile := filepath.Join(dir, "panen-old.db")
		if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
			t.Fatalf("create old backup: %v", err)
		}
		oldTime := time.Now().AddDate(0, 0, -10)
		if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
			t.Fatalf("chtimes: %v", err)
		}

		// Create a recent backup with ModTime 1 day ago
		recentFile := filepath.Join(dir, "panen-recent.db")
		if err := os.WriteFile(recentFile, []byte("recent"), 0o644); err != nil {
			t.Fatalf("create recent backup: %v", err)
		}
		recentTime := time.Now().AddDate(0, 0, -1)
		if err := os.Chtimes(recentFile, recentTime, recentTime); err != nil {
			t.Fatalf("chtimes: %v", err)
		}

		if err := svc.Cleanup(dir, 7); err != nil {
			t.Fatalf("Cleanup() error = %v", err)
		}

		if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
			t.Error("old backup was not removed")
		}
		if _, err := os.Stat(recentFile); err != nil {
			t.Error("recent backup was incorrectly removed")
		}
	})

	t.Run("handles missing backup dir", func(t *testing.T) {
		if err := svc.Cleanup("/nonexistent/path", 7); err != nil {
			t.Fatalf("Cleanup() error = %v", err)
		}
	})
}

func TestCreateBeforeDestructive(t *testing.T) {
	svc := NewBackupService()

	t.Run("creates pre-destructive backup", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := createTestDB(t, dir)
		backupDir := filepath.Join(dir, "backups")

		if err := svc.CreateBeforeDestructive(dbPath, backupDir, "delete"); err != nil {
			t.Fatalf("CreateBeforeDestructive() error = %v", err)
		}

		expected := "panen-" + time.Now().Format(time.DateOnly) + "-pre-delete.db"
		if _, err := os.Stat(filepath.Join(backupDir, expected)); err != nil {
			t.Errorf("expected backup file %s not found", expected)
		}
	})
}

func TestCreateManualBackup(t *testing.T) {
	svc := NewBackupService()

	t.Run("creates manual backup", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := createTestDB(t, dir)
		backupDir := filepath.Join(dir, "backups")

		if err := svc.CreateManualBackup(dbPath, backupDir); err != nil {
			t.Fatalf("CreateManualBackup() error = %v", err)
		}

		expected := "panen-" + time.Now().Format(time.DateOnly) + "-manual.db"
		if _, err := os.Stat(filepath.Join(backupDir, expected)); err != nil {
			t.Errorf("expected backup file %s not found", expected)
		}
	})

	t.Run("appends suffix when file exists", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := createTestDB(t, dir)
		backupDir := filepath.Join(dir, "backups")
		if err := os.MkdirAll(backupDir, 0o750); err != nil {
			t.Fatalf("create backup dir: %v", err)
		}

		// Pre-create the manual backup
		existing := filepath.Join(backupDir, "panen-"+time.Now().Format(time.DateOnly)+"-manual.db")
		if err := os.WriteFile(existing, []byte("existing"), 0o644); err != nil {
			t.Fatalf("create existing backup: %v", err)
		}

		if err := svc.CreateManualBackup(dbPath, backupDir); err != nil {
			t.Fatalf("CreateManualBackup() error = %v", err)
		}

		suffixed := "panen-" + time.Now().Format(time.DateOnly) + "-manual-1.db"
		if _, err := os.Stat(filepath.Join(backupDir, suffixed)); err != nil {
			t.Errorf("expected suffixed backup file %s not found", suffixed)
		}
	})
}

func TestListBackups(t *testing.T) {
	svc := NewBackupService()

	t.Run("lists backups sorted newest first", func(t *testing.T) {
		dir := t.TempDir()

		names := []string{"panen-2024-01-01.db", "panen-2024-01-03.db", "panen-2024-01-02.db", "not-a-backup.txt"}
		ages := []int{3, 1, 2, 0} // days ago

		for i, name := range names {
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
				t.Fatalf("create file %s: %v", name, err)
			}
			modTime := time.Now().AddDate(0, 0, -ages[i])
			if err := os.Chtimes(path, modTime, modTime); err != nil {
				t.Fatalf("chtimes: %v", err)
			}
		}

		backups, err := svc.ListBackups(dir)
		if err != nil {
			t.Fatalf("ListBackups() error = %v", err)
		}

		if len(backups) != 3 {
			t.Fatalf("ListBackups() returned %d backups, want 3", len(backups))
		}
		if backups[0].Filename != "panen-2024-01-03.db" {
			t.Errorf("first backup = %q, want panen-2024-01-03.db", backups[0].Filename)
		}
		if backups[2].Filename != "panen-2024-01-01.db" {
			t.Errorf("last backup = %q, want panen-2024-01-01.db", backups[2].Filename)
		}
	})

	t.Run("returns nil for missing directory", func(t *testing.T) {
		backups, err := svc.ListBackups("/nonexistent/path")
		if err != nil {
			t.Fatalf("ListBackups() error = %v", err)
		}
		if backups != nil {
			t.Errorf("ListBackups() = %v, want nil", backups)
		}
	})
}
