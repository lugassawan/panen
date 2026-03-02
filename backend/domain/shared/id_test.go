package shared

import (
	"regexp"
	"testing"
)

func TestNewID(t *testing.T) {
	uuidRe := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	t.Run("format", func(t *testing.T) {
		id := NewID()
		if !uuidRe.MatchString(id) {
			t.Errorf("NewID() = %q, want UUID v4 format", id)
		}
	})

	t.Run("uniqueness", func(t *testing.T) {
		seen := make(map[string]bool)
		for range 1000 {
			id := NewID()
			if seen[id] {
				t.Fatalf("NewID() produced duplicate: %s", id)
			}
			seen[id] = true
		}
	})
}
