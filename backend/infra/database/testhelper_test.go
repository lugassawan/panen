package database

import (
	"context"
	"database/sql"
	"testing"
)

// newTestDB opens an in-memory SQLite database, applies all migrations,
// and registers cleanup with t.Cleanup.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := Migrate(context.Background(), db.Conn()); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db.Conn()
}
