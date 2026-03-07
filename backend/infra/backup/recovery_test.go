package backup

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createValidDB(t *testing.T, path string) {
	t.Helper()
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, err = conn.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	conn.Close()
}

func TestTryRecoverHealthyDB(t *testing.T) {
	dir := t.TempDir()
	createValidDB(t, filepath.Join(dir, "panen.db"))

	restored, err := TryRecover(dir, filepath.Join(dir, "backups"))
	if err != nil {
		t.Fatalf("TryRecover() error = %v", err)
	}
	if restored != "" {
		t.Errorf("TryRecover() = %q, want empty string", restored)
	}
}

func TestTryRecoverMissingDB(t *testing.T) {
	dir := t.TempDir()
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		t.Fatalf("create backup dir: %v", err)
	}

	backupFile := "panen-2024-01-15.db"
	createValidDB(t, filepath.Join(backupDir, backupFile))

	restored, err := TryRecover(dir, backupDir)
	if err != nil {
		t.Fatalf("TryRecover() error = %v", err)
	}
	if restored != backupFile {
		t.Errorf("TryRecover() = %q, want %q", restored, backupFile)
	}
	if _, err := os.Stat(filepath.Join(dir, "panen.db")); err != nil {
		t.Error("restored panen.db not found")
	}
}

func TestTryRecoverCorruptDB(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "panen.db"), []byte("corrupt"), 0o644); err != nil {
		t.Fatalf("write corrupt db: %v", err)
	}

	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		t.Fatalf("create backup dir: %v", err)
	}

	backupFile := "panen-2024-01-15.db"
	createValidDB(t, filepath.Join(backupDir, backupFile))

	restored, err := TryRecover(dir, backupDir)
	if err != nil {
		t.Fatalf("TryRecover() error = %v", err)
	}
	if restored != backupFile {
		t.Errorf("TryRecover() = %q, want %q", restored, backupFile)
	}
}

func TestTryRecoverSkipsCorruptBackup(t *testing.T) {
	dir := t.TempDir()
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		t.Fatalf("create backup dir: %v", err)
	}

	// Corrupt newer backup
	if err := os.WriteFile(filepath.Join(backupDir, "panen-2024-01-16.db"), []byte("corrupt"), 0o644); err != nil {
		t.Fatalf("write corrupt backup: %v", err)
	}

	// Valid older backup
	olderBackup := "panen-2024-01-15.db"
	createValidDB(t, filepath.Join(backupDir, olderBackup))

	restored, err := TryRecover(dir, backupDir)
	if err != nil {
		t.Fatalf("TryRecover() error = %v", err)
	}
	if restored != olderBackup {
		t.Errorf("TryRecover() = %q, want %q", restored, olderBackup)
	}
}

func TestTryRecoverNoBackups(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "panen.db")
	if err := os.WriteFile(dbPath, []byte("corrupt"), 0o644); err != nil {
		t.Fatalf("write corrupt db: %v", err)
	}

	restored, err := TryRecover(dir, filepath.Join(dir, "backups"))
	if err != nil {
		t.Fatalf("TryRecover() error = %v", err)
	}
	if restored != "" {
		t.Errorf("TryRecover() = %q, want empty string", restored)
	}
	// Original file should be gone (renamed to .corrupt.*)
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Error("corrupt panen.db was not renamed")
	}

	// A .corrupt.* file should exist
	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "panen.db.corrupt.") {
			found = true
			break
		}
	}
	if !found {
		t.Error("corrupt file was not preserved as panen.db.corrupt.*")
	}
}

func TestTryRecoverAllCorruptBackups(t *testing.T) {
	dir := t.TempDir()
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		t.Fatalf("create backup dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(backupDir, "panen-2024-01-15.db"), []byte("corrupt1"), 0o644); err != nil {
		t.Fatalf("write corrupt backup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "panen-2024-01-14.db"), []byte("corrupt2"), 0o644); err != nil {
		t.Fatalf("write corrupt backup: %v", err)
	}

	restored, err := TryRecover(dir, backupDir)
	if err != nil {
		t.Fatalf("TryRecover() error = %v", err)
	}
	if restored != "" {
		t.Errorf("TryRecover() = %q, want empty string", restored)
	}
}
