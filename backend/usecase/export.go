package usecase

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/lugassawan/panen/backend/infra/backup"
)

const exportMetaName = "meta.json"

// ExportMeta holds metadata written into the export archive.
type ExportMeta struct {
	AppVersion string `json:"appVersion"`
	ExportedAt string `json:"exportedAt"`
	Checksum   string `json:"checksum"`
}

// ExportService handles data export operations.
type ExportService struct {
	dbPath     string
	appVersion string
	checkpoint func(dbPath string) error
}

// DefaultExportFilename returns a default filename for export archives.
func DefaultExportFilename() string {
	return "panen-export-" + time.Now().Format("2006-01-02") + ".zip"
}

// NewExportService creates a new ExportService.
func NewExportService(
	dbPath, appVersion string, checkpoint func(string) error,
) *ExportService {
	return &ExportService{
		dbPath:     dbPath,
		appVersion: appVersion,
		checkpoint: checkpoint,
	}
}

// Export creates a zip archive at dst containing the database and metadata.
// It returns the SHA-256 checksum of the database file included in the archive.
func (s *ExportService) Export(dst string) (string, error) {
	if err := s.checkpoint(s.dbPath); err != nil {
		return "", fmt.Errorf("checkpoint before export: %w", err)
	}

	checksum, err := fileChecksum(s.dbPath)
	if err != nil {
		return "", fmt.Errorf("compute checksum: %w", err)
	}

	meta := ExportMeta{
		AppVersion: s.appVersion,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Checksum:   checksum,
	}

	if err := writeExportZip(dst, s.dbPath, meta); err != nil {
		_ = os.Remove(dst)
		return "", fmt.Errorf("write export zip: %w", err)
	}

	return checksum, nil
}

func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeExportZip(dst, dbPath string, meta ExportMeta) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	if err := addFileToZip(zw, backup.DBFilename, dbPath); err != nil {
		return fmt.Errorf("add database to zip: %w", err)
	}

	metaJSON, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	w, err := zw.Create(exportMetaName)
	if err != nil {
		return fmt.Errorf("create meta entry: %w", err)
	}
	if _, err := w.Write(metaJSON); err != nil {
		return fmt.Errorf("write meta entry: %w", err)
	}

	return nil
}

func addFileToZip(zw *zip.Writer, name, srcPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	header.Method = zip.Deflate

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, src)
	return err
}
