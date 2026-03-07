//go:build darwin

package updater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type darwinInstaller struct{}

// NewPlatformInstaller returns a macOS-specific installer.
func NewPlatformInstaller() PlatformInstaller {
	return &darwinInstaller{}
}

func (i *darwinInstaller) ArchiveName() string {
	return "panen-darwin-universal.zip"
}

// InstallPath walks up from the current executable to find the .app bundle root.
func (i *darwinInstaller) InstallPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("eval symlinks: %w", err)
	}

	// Walk up to find .app bundle (e.g. /Applications/panen.app/Contents/MacOS/panen)
	dir := exe
	for {
		dir = filepath.Dir(dir)
		if dir == "/" || dir == "." {
			break
		}
		if strings.HasSuffix(dir, ".app") {
			return dir, nil
		}
	}
	return "", fmt.Errorf("no .app bundle found for %s", exe)
}

// Install moves the existing .app to .app.backup and places the new one.
func (i *darwinInstaller) Install(
	extractedDir, installPath string,
) error {
	backupPath := installPath + ".backup"

	if err := os.Rename(installPath, backupPath); err != nil {
		return fmt.Errorf("backup current app: %w", err)
	}

	// Find the .app in the extracted directory
	entries, err := os.ReadDir(extractedDir)
	if err != nil {
		return fmt.Errorf("read extracted dir: %w", err)
	}
	var appDir string
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".app") {
			appDir = filepath.Join(extractedDir, e.Name())
			break
		}
	}
	if appDir == "" {
		return errors.New("no .app found in extracted archive")
	}

	if err := os.Rename(appDir, installPath); err != nil {
		return fmt.Errorf("install new app: %w", err)
	}
	return nil
}

func (i *darwinInstaller) Rollback(installPath string) error {
	backupPath := installPath + ".backup"
	_ = os.RemoveAll(installPath)
	if err := os.Rename(backupPath, installPath); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}
	return nil
}

func (i *darwinInstaller) CleanupBackup(installPath string) error {
	backupPath := installPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(backupPath)
}
