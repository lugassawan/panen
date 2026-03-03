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

func TestMigrateV2AddsBrokerColumns(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO user_profiles (id, name, created_at, updated_at)
		VALUES ('u1', 'Test', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO brokerage_accounts
		(id, profile_id, broker_name, created_at, updated_at)
		VALUES ('b1', 'u1', 'Test Broker', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert brokerage: %v", err)
	}

	var sellTaxPct float64
	var brokerCode string
	err = db.QueryRowContext(ctx,
		"SELECT sell_tax_pct, broker_code FROM brokerage_accounts WHERE id = 'b1'",
	).Scan(&sellTaxPct, &brokerCode)
	if err != nil {
		t.Fatalf("query error = %v", err)
	}
	if sellTaxPct != 0 {
		t.Errorf("sell_tax_pct = %v, want 0", sellTaxPct)
	}
	if brokerCode != "" {
		t.Errorf("broker_code = %q, want empty string", brokerCode)
	}
}
