package platform

import (
	"os"
	"path/filepath"
)

// DataDir returns the platform-appropriate directory for Panen's data files.
//
// The app directory name varies by build:
//   - Production: Panen
//   - Dev (wails dev): Panen-Dev
//
// Results by platform (production):
//   - macOS:   ~/Library/Application Support/Panen/data
//   - Linux:   ~/.config/Panen/data  (or $XDG_CONFIG_HOME/Panen/data)
//   - Windows: %APPDATA%\Panen\data
//
// The directory is NOT created — the caller is responsible for calling
// os.MkdirAll before writing files.
func DataDir() (string, error) {
	base, err := appBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "data"), nil
}

// LogDir returns the platform-appropriate directory for Panen's log files.
//
// Results by platform (production):
//   - macOS:   ~/Library/Application Support/Panen/logs
//   - Linux:   ~/.config/Panen/logs  (or $XDG_CONFIG_HOME/Panen/logs)
//   - Windows: %APPDATA%\Panen\logs
//
// The directory is NOT created — the caller is responsible for calling
// os.MkdirAll before writing files.
func LogDir() (string, error) {
	base, err := appBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "logs"), nil
}

// BackupDir returns the platform-appropriate directory for Panen's database backups.
//
// Results by platform (production):
//   - macOS:   ~/Library/Application Support/Panen/backups
//   - Linux:   ~/.config/Panen/backups  (or $XDG_CONFIG_HOME/Panen/backups)
//   - Windows: %APPDATA%\Panen\backups
//
// The directory is NOT created — the caller is responsible for calling
// os.MkdirAll before writing files.
func BackupDir() (string, error) {
	base, err := appBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "backups"), nil
}

// appBaseDir returns the platform-appropriate base directory for Panen.
func appBaseDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, appName), nil
}
