//go:build windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type windowsInstaller struct{}

// NewPlatformInstaller returns a Windows-specific installer.
func NewPlatformInstaller() PlatformInstaller {
	return &windowsInstaller{}
}

func (i *windowsInstaller) ArchiveName() string {
	return "panen-windows-amd64.zip"
}

func (i *windowsInstaller) InstallPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	return filepath.EvalSymlinks(exe)
}

func (i *windowsInstaller) Install(
	extractedDir, installPath string,
) error {
	backupPath := installPath + ".backup"

	// Windows allows renaming a running exe
	if err := os.Rename(installPath, backupPath); err != nil {
		return fmt.Errorf("backup current exe: %w", err)
	}

	newExe := filepath.Join(extractedDir, "panen.exe")
	if err := copyFileWin(newExe, installPath); err != nil {
		return fmt.Errorf("install new exe: %w", err)
	}
	return nil
}

func (i *windowsInstaller) Rollback(installPath string) error {
	backupPath := installPath + ".backup"
	_ = os.Remove(installPath)
	if err := os.Rename(backupPath, installPath); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}
	return nil
}

func (i *windowsInstaller) CleanupBackup(installPath string) error {
	backupPath := installPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}
	// Retry cleanup since previous process may have just exited
	var lastErr error
	for range 3 {
		if err := os.Remove(backupPath); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return lastErr
}

func copyFileWin(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}
