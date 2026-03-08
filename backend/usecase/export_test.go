package usecase

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func createTestDB(t *testing.T, dir string) string {
	t.Helper()
	dbPath := filepath.Join(dir, "panen.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	_, err = conn.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY); INSERT INTO test VALUES (1)")
	if err != nil {
		t.Fatalf("create test table: %v", err)
	}
	conn.Close()
	return dbPath
}

func noopCheckpoint(_ string) error { return nil }

func TestExport(t *testing.T) {
	t.Run("creates valid zip with db and meta", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := createTestDB(t, dir)
		dst := filepath.Join(dir, "export.zip")

		svc := NewExportService(dbPath, "1.0.0", noopCheckpoint)
		checksum, err := svc.Export(dst)
		if err != nil {
			t.Fatalf("Export() error = %v", err)
		}
		if checksum == "" {
			t.Fatal("Export() returned empty checksum")
		}

		zr, err := zip.OpenReader(dst)
		if err != nil {
			t.Fatalf("open zip: %v", err)
		}
		defer zr.Close()

		var hasDB, hasMeta bool
		for _, f := range zr.File {
			switch f.Name {
			case "panen.db":
				hasDB = true
			case "meta.json":
				hasMeta = true
				rc, err := f.Open()
				if err != nil {
					t.Fatalf("open meta.json: %v", err)
				}
				var meta ExportMeta
				if err := json.NewDecoder(rc).Decode(&meta); err != nil {
					t.Fatalf("decode meta.json: %v", err)
				}
				rc.Close()

				if meta.AppVersion != "1.0.0" {
					t.Errorf("meta.AppVersion = %q, want %q", meta.AppVersion, "1.0.0")
				}
				if meta.Checksum != checksum {
					t.Errorf("meta.Checksum = %q, want %q", meta.Checksum, checksum)
				}
				if meta.ExportedAt == "" {
					t.Error("meta.ExportedAt is empty")
				}
			}
		}

		if !hasDB {
			t.Error("zip missing panen.db")
		}
		if !hasMeta {
			t.Error("zip missing meta.json")
		}
	})

	t.Run("returns error for missing db", func(t *testing.T) {
		dir := t.TempDir()
		dst := filepath.Join(dir, "export.zip")

		svc := NewExportService(filepath.Join(dir, "nonexistent.db"), "1.0.0", noopCheckpoint)
		_, err := svc.Export(dst)
		if err == nil {
			t.Fatal("Export() expected error for missing db")
		}
	})
}

func TestDefaultExportFilename(t *testing.T) {
	name := DefaultExportFilename()
	if name == "" {
		t.Error("DefaultExportFilename() returned empty string")
	}
}
