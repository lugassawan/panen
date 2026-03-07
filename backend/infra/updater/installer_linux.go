//go:build linux

package updater

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type linuxInstaller struct{}

// NewPlatformInstaller returns a Linux-specific installer.
func NewPlatformInstaller() PlatformInstaller {
	return &linuxInstaller{}
}

func (i *linuxInstaller) ArchiveName() string {
	return "panen-linux-amd64.tar.gz"
}

func (i *linuxInstaller) InstallPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	return filepath.EvalSymlinks(exe)
}

func (i *linuxInstaller) Install(
	extractedDir, installPath string,
) error {
	backupPath := installPath + ".backup"

	if err := os.Rename(installPath, backupPath); err != nil {
		return fmt.Errorf("backup current binary: %w", err)
	}

	newBin := filepath.Join(extractedDir, "panen")
	if err := copyFile(newBin, installPath, 0o755); err != nil {
		return fmt.Errorf("install new binary: %w", err)
	}

	// Update desktop file and icon if present
	i.updateDesktopAssets(extractedDir)

	return nil
}

func (i *linuxInstaller) Rollback(installPath string) error {
	backupPath := installPath + ".backup"
	_ = os.Remove(installPath)
	if err := os.Rename(backupPath, installPath); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}
	return nil
}

func (i *linuxInstaller) CleanupBackup(installPath string) error {
	backupPath := installPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(backupPath)
}

func (i *linuxInstaller) updateDesktopAssets(extractedDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// Desktop file
	srcDesktop := filepath.Join(extractedDir, "panen.desktop")
	if _, err := os.Stat(srcDesktop); err == nil {
		dstDesktop := filepath.Join(
			home, ".local", "share", "applications", "panen.desktop",
		)
		_ = copyFile(srcDesktop, dstDesktop, 0o644)
	}

	// Icon
	srcIcon := filepath.Join(extractedDir, "panen.png")
	if _, err := os.Stat(srcIcon); err == nil {
		dstIcon := filepath.Join(
			home, ".local", "share", "icons", "panen.png",
		)
		_ = copyFile(srcIcon, dstIcon, 0o644)
	}
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	out, err := os.OpenFile(
		dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm,
	)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
