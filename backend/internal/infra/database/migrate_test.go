package database

import (
	"context"
	"testing"
)

func TestMigrate(t *testing.T) {
	t.Run("creates schema_migrations table", func(t *testing.T) {
		db := newTestDB(t)

		var count int
		err := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'",
		).Scan(&count)
		if err != nil {
			t.Fatalf("query error = %v", err)
		}
		if count != 1 {
			t.Errorf("schema_migrations table count = %d, want 1", count)
		}
	})

	t.Run("records migration version", func(t *testing.T) {
		db := newTestDB(t)

		var version int
		err := db.QueryRowContext(context.Background(),
			"SELECT MAX(version) FROM schema_migrations",
		).Scan(&version)
		if err != nil {
			t.Fatalf("query error = %v", err)
		}
		if version != len(migrations) {
			t.Errorf("version = %d, want %d", version, len(migrations))
		}
	})

	t.Run("creates all tables", func(t *testing.T) {
		db := newTestDB(t)

		tables := []string{
			"user_profiles",
			"brokerage_accounts",
			"portfolios",
			"holdings",
			"buy_transactions",
			"stock_data",
		}
		for _, table := range tables {
			var count int
			err := db.QueryRowContext(context.Background(),
				"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table,
			).Scan(&count)
			if err != nil {
				t.Fatalf("query error for %s: %v", table, err)
			}
			if count != 1 {
				t.Errorf("table %s not found", table)
			}
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		db := newTestDB(t)

		// Running migrate again should be a no-op.
		if err := Migrate(context.Background(), db); err != nil {
			t.Fatalf("second Migrate() error = %v", err)
		}

		var version int
		err := db.QueryRowContext(context.Background(),
			"SELECT MAX(version) FROM schema_migrations",
		).Scan(&version)
		if err != nil {
			t.Fatalf("query error = %v", err)
		}
		if version != len(migrations) {
			t.Errorf("version = %d after second migrate, want %d", version, len(migrations))
		}
	})
}
