package platform

import (
	"os"
	"os/user"
	"runtime"
)

// Username returns the display name of the current OS user.
//
// It tries, in order:
//  1. The display name from os/user (GECOS on Unix, full name on macOS/Windows)
//  2. The login name from os/user
//  3. $USER (Unix) or $USERNAME (Windows)
//  4. "Default" as a final fallback
func Username() string {
	if u, err := user.Current(); err == nil {
		if u.Name != "" {
			return u.Name
		}
		if u.Username != "" {
			return u.Username
		}
	}

	envVar := "USER"
	if runtime.GOOS == "windows" {
		envVar = "USERNAME"
	}
	if name := os.Getenv(envVar); name != "" {
		return name
	}

	return "Default"
}
