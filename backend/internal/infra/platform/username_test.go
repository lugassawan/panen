package platform

import (
	"os/user"
	"testing"
)

func TestUsername(t *testing.T) {
	t.Run("returns non-empty string", func(t *testing.T) {
		name := Username()
		if name == "" {
			t.Fatal("Username() returned empty string")
		}
	})

	t.Run("matches os/user when available", func(t *testing.T) {
		u, err := user.Current()
		if err != nil {
			t.Skip("os/user.Current() unavailable:", err)
		}

		got := Username()

		// Username should return the display name or login name from os/user.
		if got != u.Name && got != u.Username {
			t.Errorf("Username() = %q, want %q or %q", got, u.Name, u.Username)
		}
	})

	t.Run("does not return Default when os user is available", func(t *testing.T) {
		_, err := user.Current()
		if err != nil {
			t.Skip("os/user.Current() unavailable:", err)
		}

		got := Username()
		if got == "Default" {
			t.Error("Username() = \"Default\", want actual OS username")
		}
	})
}
