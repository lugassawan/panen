package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogDir(t *testing.T) {
	testDirFunc(t, "LogDir", LogDir, "logs")
}

func TestDataDir(t *testing.T) {
	testDirFunc(t, "DataDir", DataDir, "data")
}

func TestBackupDir(t *testing.T) {
	testDirFunc(t, "BackupDir", BackupDir, "backups")
}

func testDirFunc(t *testing.T, name string, fn func() (string, error), suffix string) {
	t.Helper()

	t.Run("returns non-empty path", func(t *testing.T) {
		dir, err := fn()
		if err != nil {
			t.Fatalf("%s() error = %v", name, err)
		}
		if dir == "" {
			t.Fatalf("%s() returned empty string", name)
		}
	})

	t.Run("ends with appName/"+suffix, func(t *testing.T) {
		dir, err := fn()
		if err != nil {
			t.Fatalf("%s() error = %v", name, err)
		}
		want := filepath.Join(appName, suffix)
		if !strings.HasSuffix(dir, want) {
			t.Errorf("%s() = %q, want suffix %q", name, dir, want)
		}
	})

	t.Run("is rooted under UserConfigDir", func(t *testing.T) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			t.Fatalf("os.UserConfigDir() error = %v", err)
		}
		dir, err := fn()
		if err != nil {
			t.Fatalf("%s() error = %v", name, err)
		}
		if !strings.HasPrefix(dir, configDir) {
			t.Errorf("%s() = %q, want prefix %q", name, dir, configDir)
		}
	})
}
