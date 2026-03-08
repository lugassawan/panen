package usecase

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func createTestExport(t *testing.T, dir, version, checksum string) string {
	t.Helper()
	archivePath := filepath.Join(dir, "test-export.zip")
	dbPath := createTestDB(t, dir)

	if checksum == "" {
		var err error
		checksum, err = fileChecksum(dbPath)
		if err != nil {
			t.Fatalf("compute checksum: %v", err)
		}
	}

	meta := ExportMeta{
		AppVersion: version,
		ExportedAt: "2026-01-01T00:00:00Z",
		Checksum:   checksum,
	}

	if err := writeExportZip(archivePath, dbPath, meta); err != nil {
		t.Fatalf("create test export: %v", err)
	}
	return archivePath
}

func noopBackup(_, _, _ string) error { return nil }

func TestImportPreview(t *testing.T) {
	t.Run("returns metadata from archive", func(t *testing.T) {
		dir := t.TempDir()
		archivePath := createTestExport(t, dir, "0.9.0", "")

		svc := NewImportService(
			filepath.Join(dir, "live.db"), filepath.Join(dir, "backups"),
			"1.0.0", noopBackup,
		)

		result, err := svc.Preview(archivePath)
		if err != nil {
			t.Fatalf("Preview() error = %v", err)
		}

		if result.AppVersion != "0.9.0" {
			t.Errorf("AppVersion = %q, want %q", result.AppVersion, "0.9.0")
		}
		if result.ExportedAt != "2026-01-01T00:00:00Z" {
			t.Errorf("ExportedAt = %q, want %q", result.ExportedAt, "2026-01-01T00:00:00Z")
		}
		if result.DbSize <= 0 {
			t.Errorf("DbSize = %d, want > 0", result.DbSize)
		}
		if result.Checksum == "" {
			t.Error("Checksum is empty")
		}
	})

	t.Run("returns error for invalid archive", func(t *testing.T) {
		dir := t.TempDir()
		badFile := filepath.Join(dir, "bad.zip")
		if err := os.WriteFile(badFile, []byte("not a zip"), 0o644); err != nil {
			t.Fatalf("write bad file: %v", err)
		}

		svc := NewImportService(
			filepath.Join(dir, "live.db"), filepath.Join(dir, "backups"),
			"1.0.0", noopBackup,
		)

		_, err := svc.Preview(badFile)
		if err == nil {
			t.Fatal("Preview() expected error for invalid archive")
		}
	})

	t.Run("returns error for archive missing meta", func(t *testing.T) {
		dir := t.TempDir()
		archivePath := filepath.Join(dir, "no-meta.zip")
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create file: %v", err)
		}
		zw := zip.NewWriter(f)
		w, _ := zw.Create("panen.db")
		_, _ = w.Write([]byte("data"))
		zw.Close()
		f.Close()

		svc := NewImportService(
			filepath.Join(dir, "live.db"), filepath.Join(dir, "backups"),
			"1.0.0", noopBackup,
		)

		_, err = svc.Preview(archivePath)
		if err == nil {
			t.Fatal("Preview() expected error for archive missing meta.json")
		}
	})
}

func TestImport(t *testing.T) {
	t.Run("replaces database from archive", func(t *testing.T) {
		dir := t.TempDir()
		archivePath := createTestExport(t, dir, "1.0.0", "")

		liveDBPath := filepath.Join(dir, "live.db")
		if err := os.WriteFile(liveDBPath, []byte("old data"), 0o644); err != nil {
			t.Fatalf("create live db: %v", err)
		}
		backupDir := filepath.Join(dir, "backups")

		svc := NewImportService(liveDBPath, backupDir, "1.0.0", noopBackup)
		if err := svc.Import(archivePath); err != nil {
			t.Fatalf("Import() error = %v", err)
		}

		data, err := os.ReadFile(liveDBPath)
		if err != nil {
			t.Fatalf("read live db: %v", err)
		}
		if string(data) == "old data" {
			t.Error("Import() did not replace the database")
		}
	})

	t.Run("fails on checksum mismatch", func(t *testing.T) {
		dir := t.TempDir()
		archivePath := createTestExport(t, dir, "1.0.0", "bad-checksum")

		liveDBPath := filepath.Join(dir, "live.db")
		if err := os.WriteFile(liveDBPath, []byte("old data"), 0o644); err != nil {
			t.Fatalf("create live db: %v", err)
		}
		backupDir := filepath.Join(dir, "backups")

		svc := NewImportService(liveDBPath, backupDir, "1.0.0", noopBackup)
		err := svc.Import(archivePath)
		if err == nil {
			t.Fatal("Import() expected error for checksum mismatch")
		}
	})

	t.Run("calls backup before replacing", func(t *testing.T) {
		dir := t.TempDir()
		archivePath := createTestExport(t, dir, "1.0.0", "")

		liveDBPath := filepath.Join(dir, "live.db")
		if err := os.WriteFile(liveDBPath, []byte("old data"), 0o644); err != nil {
			t.Fatalf("create live db: %v", err)
		}
		backupDir := filepath.Join(dir, "backups")

		backupCalled := false
		backupFn := func(_, _, reason string) error {
			backupCalled = true
			if reason != importBackupReason {
				t.Errorf("backup reason = %q, want %q", reason, importBackupReason)
			}
			return nil
		}

		svc := NewImportService(liveDBPath, backupDir, "1.0.0", backupFn)
		if err := svc.Import(archivePath); err != nil {
			t.Fatalf("Import() error = %v", err)
		}

		if !backupCalled {
			t.Error("Import() did not call backup before replacing")
		}
	})
}

func TestImportMissingDB(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "no-db.zip")

	// Create a zip with only meta.json
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("meta.json")
	meta := ExportMeta{AppVersion: "1.0.0", ExportedAt: "2026-01-01T00:00:00Z"}
	data, _ := json.Marshal(meta)
	_, _ = w.Write(data)
	zw.Close()
	f.Close()

	svc := NewImportService(
		filepath.Join(dir, "live.db"), filepath.Join(dir, "backups"),
		"1.0.0", noopBackup,
	)

	_, err = svc.Preview(archivePath)
	if err == nil {
		t.Fatal("Preview() expected error for archive missing panen.db")
	}

	err = svc.Import(archivePath)
	if err == nil {
		t.Fatal("Import() expected error for archive missing panen.db")
	}
}
