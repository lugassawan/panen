package usecase

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/lugassawan/panen/backend/infra/backup"
	"github.com/lugassawan/panen/backend/infra/database"

	_ "modernc.org/sqlite"
)

const (
	importBackupReason = "import"
	maxImportDBSize    = 1 << 30 // 1 GB
)

// ImportPreviewResult holds metadata previewed from an import archive.
type ImportPreviewResult struct {
	AppVersion string
	ExportedAt string
	Checksum   string
	DbSize     int64
}

// ImportService handles data import operations.
type ImportService struct {
	dbPath     string
	backupDir  string
	appVersion string
	backupFn   func(dbPath, backupDir, reason string) error
}

// NewImportService creates a new ImportService.
func NewImportService(
	dbPath, backupDir, appVersion string,
	backupFn func(string, string, string) error,
) *ImportService {
	return &ImportService{
		dbPath:     dbPath,
		backupDir:  backupDir,
		appVersion: appVersion,
		backupFn:   backupFn,
	}
}

// Preview reads the import archive and returns metadata without modifying anything.
func (s *ImportService) Preview(archivePath string) (*ImportPreviewResult, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer zr.Close()

	meta, err := readMeta(zr)
	if err != nil {
		return nil, err
	}

	dbFile := findZipEntry(zr, backup.DBFilename)
	if dbFile == nil {
		return nil, fmt.Errorf("archive missing %s", backup.DBFilename)
	}

	return &ImportPreviewResult{
		AppVersion: meta.AppVersion,
		ExportedAt: meta.ExportedAt,
		Checksum:   meta.Checksum,
		DbSize:     int64(dbFile.UncompressedSize64), //nolint:gosec // DB files won't exceed int64 max
	}, nil
}

// Import replaces the current database with the one from the archive.
// It creates a safety backup first and verifies the checksum.
func (s *ImportService) Import(archivePath string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer zr.Close()

	meta, err := readMeta(zr)
	if err != nil {
		return err
	}

	dbEntry := findZipEntry(zr, backup.DBFilename)
	if dbEntry == nil {
		return fmt.Errorf("archive missing %s", backup.DBFilename)
	}

	// Extract to temp file first to verify checksum before replacing.
	tmpPath, err := extractDBToTemp(dbEntry)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	if meta.Checksum != "" {
		actual, checksumErr := fileChecksum(tmpPath)
		if checksumErr != nil {
			return fmt.Errorf("verify checksum: %w", checksumErr)
		}
		if actual != meta.Checksum {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", meta.Checksum, actual)
		}
	}

	if err := quickCheckDB(tmpPath); err != nil {
		return fmt.Errorf("imported database validation: %w", err)
	}

	// Safety backup before replacing.
	if err := s.backupFn(s.dbPath, s.backupDir, importBackupReason); err != nil {
		return fmt.Errorf("pre-import backup: %w", err)
	}

	// Replace the database file.
	if err := replaceFile(tmpPath, s.dbPath); err != nil {
		return fmt.Errorf("replace database: %w", err)
	}

	return nil
}

func readMeta(zr *zip.ReadCloser) (*ExportMeta, error) {
	metaEntry := findZipEntry(zr, exportMetaName)
	if metaEntry == nil {
		return nil, fmt.Errorf("archive missing %s", exportMetaName)
	}

	rc, err := metaEntry.Open()
	if err != nil {
		return nil, fmt.Errorf("open meta.json: %w", err)
	}
	defer rc.Close()

	var meta ExportMeta
	if err := json.NewDecoder(rc).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decode meta.json: %w", err)
	}
	return &meta, nil
}

func findZipEntry(zr *zip.ReadCloser, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func extractDBToTemp(entry *zip.File) (string, error) {
	rc, err := entry.Open()
	if err != nil {
		return "", fmt.Errorf("open db entry: %w", err)
	}
	defer rc.Close()

	tmp, err := os.CreateTemp("", "panen-import-*.db")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	limited := io.LimitReader(rc, maxImportDBSize+1)
	written, err := io.Copy(tmp, limited) //nolint:gosec // zip bomb mitigated by LimitReader
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name()) //nolint:gosec // temp file we just created
		return "", fmt.Errorf("extract db: %w", err)
	}
	if written > maxImportDBSize {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name()) //nolint:gosec // temp file we just created
		return "", fmt.Errorf("database exceeds maximum size (%d bytes)", maxImportDBSize)
	}

	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name()) //nolint:gosec // temp file we just created
		return "", fmt.Errorf("sync temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name()) //nolint:gosec // temp file we just created
		return "", err
	}

	return tmp.Name(), nil
}

func replaceFile(src, dst string) error {
	// Copy to a temp file next to dst first, then rename for atomic replacement.
	tmpDst := dst + ".importing"
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(tmpDst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmpDst)
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		_ = os.Remove(tmpDst)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmpDst)
		return err
	}

	// Remove stale WAL/SHM files that could conflict with the imported database.
	_ = os.Remove(dst + "-wal")
	_ = os.Remove(dst + "-shm")

	return os.Rename(tmpDst, dst)
}

// quickCheckDB opens a temporary connection and runs PRAGMA quick_check.
func quickCheckDB(dbPath string) error {
	conn, err := sql.Open(database.SQLiteDriver, dbPath+"?_pragma=busy_timeout%3d5000")
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
