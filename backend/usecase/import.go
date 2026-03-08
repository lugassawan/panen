package usecase

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const importBackupReason = "import"

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

	dbFile := findZipEntry(zr, exportDBName)
	if dbFile == nil {
		return nil, errors.New("archive missing panen.db")
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

	dbEntry := findZipEntry(zr, exportDBName)
	if dbEntry == nil {
		return errors.New("archive missing panen.db")
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
		return nil, errors.New("archive missing meta.json")
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

	if _, err := io.Copy(tmp, rc); err != nil { //nolint:gosec // trusted archive from user
		_ = tmp.Close()
		_ = os.Remove(tmp.Name()) //nolint:gosec // temp file we just created
		return "", fmt.Errorf("extract db: %w", err)
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
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
