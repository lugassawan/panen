package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDataDir(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		dir, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		if dir == "" {
			t.Fatal("DataDir() returned empty string")
		}
	})

	t.Run("ends with Panen/data", func(t *testing.T) {
		dir, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		want := filepath.Join("Panen", "data")
		if !strings.HasSuffix(dir, want) {
			t.Errorf("DataDir() = %q, want suffix %q", dir, want)
		}
	})

	t.Run("is rooted under UserConfigDir", func(t *testing.T) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			t.Fatalf("os.UserConfigDir() error = %v", err)
		}
		dir, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		if !strings.HasPrefix(dir, configDir) {
			t.Errorf("DataDir() = %q, want prefix %q", dir, configDir)
		}
	})
}
