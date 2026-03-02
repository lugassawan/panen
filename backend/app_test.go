package backend

import (
	"context"
	"testing"
)

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil {
		t.Fatal("NewApp returned nil")
	}
}

func TestShutdownNilDB(t *testing.T) {
	a := NewApp()
	// Shutdown with nil db should not panic.
	a.Shutdown(context.Background())
}
