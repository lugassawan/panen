package user

import "testing"

func TestNewProfile(t *testing.T) {
	p := NewProfile("Alice")

	if p.ID == "" {
		t.Error("expected non-empty ID")
	}
	if p.Name != "Alice" {
		t.Errorf("Name = %q, want %q", p.Name, "Alice")
	}
	if p.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
	if p.CreatedAt != p.UpdatedAt {
		t.Error("expected CreatedAt == UpdatedAt for new profile")
	}
}

func TestNewProfileGeneratesUniqueIDs(t *testing.T) {
	p1 := NewProfile("Alice")
	p2 := NewProfile("Alice")

	if p1.ID == p2.ID {
		t.Error("expected unique IDs for different profiles")
	}
}
