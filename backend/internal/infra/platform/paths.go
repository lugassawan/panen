package platform

import (
	"os"
	"path/filepath"
)

// DataDir returns the platform-appropriate directory for Panen's data files.
//
// Results by platform:
//   - macOS:   ~/Library/Application Support/Panen/data
//   - Linux:   ~/.config/panen/data  (or $XDG_CONFIG_HOME/panen/data)
//   - Windows: %APPDATA%\Panen\data
//
// The directory is NOT created — the caller is responsible for calling
// os.MkdirAll before writing files.
func DataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "Panen", "data"), nil
}
