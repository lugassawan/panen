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

// InstallPath returns the canonical install location for macOS.
func (i *darwinInstaller) InstallPath() (string, error) {
	return "/Applications/Panen.app", nil
}

// Install moves the existing .app to .app.backup and places the new one.
// If the app was running from a different location (e.g. ~/Applications),
// it cleans up the old location after installing.
func (i *darwinInstaller) Install(
	extractedDir, installPath string,
) error {
	backupPath := installPath + ".backup"

	// Back up existing install at target — ignore if not present yet
	if err := os.Rename(installPath, backupPath); err != nil && !os.IsNotExist(err) {
		legacyPath := filepath.Join(filepath.Dir(installPath), "panen.app")
		if renameErr := os.Rename(legacyPath, backupPath); renameErr != nil {
			return fmt.Errorf("backup current app: %w", err)
		}
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

	// Clean up old location if migrating
	oldPath, err := i.currentAppPath()
	if err == nil && filepath.Dir(oldPath) != filepath.Dir(installPath) {
		_ = os.RemoveAll(oldPath)
		legacyOld := filepath.Join(filepath.Dir(oldPath), "panen.app")
		_ = os.RemoveAll(legacyOld)
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

// currentAppPath walks up from the current executable to find the .app bundle root.
func (i *darwinInstaller) currentAppPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("eval symlinks: %w", err)
	}

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
