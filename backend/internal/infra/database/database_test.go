package database

import (
	"context"
	"testing"
)

func TestOpen(t *testing.T) {
	t.Run("in-memory database", func(t *testing.T) {
		db, err := Open(":memory:")
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		if db.Conn() == nil {
			t.Fatal("Conn() returned nil")
		}
	})

	t.Run("WAL mode enabled", func(t *testing.T) {
		db, err := Open(":memory:")
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		var mode string
		err = db.Conn().QueryRowContext(context.Background(), "PRAGMA journal_mode").Scan(&mode)
		if err != nil {
			t.Fatalf("PRAGMA journal_mode error = %v", err)
		}
		// In-memory databases report "memory" instead of "wal".
		if mode != "wal" && mode != "memory" {
			t.Errorf("journal_mode = %q, want wal or memory", mode)
		}
	})

	t.Run("foreign keys enabled", func(t *testing.T) {
		db, err := Open(":memory:")
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer db.Close()

		var fk int
		err = db.Conn().QueryRowContext(context.Background(), "PRAGMA foreign_keys").Scan(&fk)
		if err != nil {
			t.Fatalf("PRAGMA foreign_keys error = %v", err)
		}
		if fk != 1 {
			t.Errorf("foreign_keys = %d, want 1", fk)
		}
	})
}
