package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxExtractSize = 500 << 20 // 500 MB per file

// Extractor handles archive extraction for supported formats.
type Extractor struct{}

// Extract decompresses archivePath into destDir.
// Detects format by extension: .zip or .tar.gz/.tgz.
func (e *Extractor) Extract(archivePath, destDir string) error {
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(archivePath, destDir)
	case strings.HasSuffix(lower, ".tar.gz"),
		strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(archivePath, destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}

func extractZip(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if err := extractZipEntry(f, destDir); err != nil {
			return err
		}
	}
	return nil
}

func extractZipEntry(f *zip.File, destDir string) error {
	target, err := sanitizePath(destDir, f.Name)
	if err != nil {
		return err
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o750)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry: %w", err)
	}
	defer rc.Close()

	out, err := os.OpenFile(
		target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode(),
	)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(rc, maxExtractSize)); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func extractTarGz(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tar.gz: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}
		if err := extractTarEntry(hdr, tr, destDir); err != nil {
			return err
		}
	}
	return nil
}

func extractTarEntry(
	hdr *tar.Header,
	r io.Reader,
	destDir string,
) error {
	target, err := sanitizePath(destDir, hdr.Name)
	if err != nil {
		return err
	}

	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, 0o750) //nolint:gosec // target validated by sanitizePath
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil { //nolint:gosec // target validated by sanitizePath
			return fmt.Errorf("create parent dir: %w", err)
		}
		mode := os.FileMode(0o644)
		if hdr.Mode&0o111 != 0 {
			mode = 0o755
		}
		out, err := os.OpenFile( //nolint:gosec // target validated by sanitizePath
			target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode,
		)
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer out.Close()
		if _, err := io.Copy(out, io.LimitReader(r, maxExtractSize)); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		return nil
	default:
		// Skip symlinks, devices, etc.
		return nil
	}
}

// sanitizePath validates that name stays within destDir (zip-slip protection)
// and returns the cleaned absolute path.
func sanitizePath(destDir, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf(
			"archive entry has absolute path: %s", name,
		)
	}
	clean := filepath.Clean(name)
	// #nosec G305 — canonical zip-slip guard: verify joined path stays under destDir
	target := filepath.Join(destDir, clean)
	if !strings.HasPrefix(
		target, filepath.Clean(destDir)+string(os.PathSeparator),
	) {
		return "", fmt.Errorf(
			"archive entry escapes destination: %s", name,
		)
	}
	return target, nil
}
