package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractZip(t *testing.T) {
	src := createTestZip(t, map[string]string{
		"dir/hello.txt": "hello world",
		"root.txt":      "root content",
	})

	destDir := t.TempDir()
	e := &Extractor{}
	if err := e.Extract(src, destDir); err != nil {
		t.Fatalf("Extract zip: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "root.txt"), "root content")
	assertFileContent(
		t, filepath.Join(destDir, "dir", "hello.txt"), "hello world",
	)
}

func TestExtractTarGz(t *testing.T) {
	src := createTestTarGz(t, map[string]string{
		"bin/panen": "binary content",
	})

	destDir := t.TempDir()
	e := &Extractor{}
	if err := e.Extract(src, destDir); err != nil {
		t.Fatalf("Extract tar.gz: %v", err)
	}

	assertFileContent(
		t, filepath.Join(destDir, "bin", "panen"), "binary content",
	)
}

func TestExtractZipSlipRejection(t *testing.T) {
	src := createTestZip(t, map[string]string{
		"../escape.txt": "malicious",
	})

	destDir := t.TempDir()
	e := &Extractor{}
	err := e.Extract(src, destDir)
	if err == nil {
		t.Fatal("expected zip-slip error")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractTarGzSlipRejection(t *testing.T) {
	src := createTestTarGz(t, map[string]string{
		"../escape.txt": "malicious",
	})

	destDir := t.TempDir()
	e := &Extractor{}
	err := e.Extract(src, destDir)
	if err == nil {
		t.Fatal("expected zip-slip error")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractUnsupportedFormat(t *testing.T) {
	e := &Extractor{}
	err := e.Extract("/tmp/test.rar", t.TempDir())
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported archive format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"normal file", "foo/bar.txt", false},
		{"root file", "file.txt", false},
		{"absolute path", "/etc/passwd", true},
		{"parent traversal", "../escape", true},
		{"nested traversal", "foo/../../escape", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sanitizePath("/dest", tc.path)
			if tc.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helpers

func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func createTestTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf(
			"%s content = %q, want %q", filepath.Base(path), got, want,
		)
	}
}
