package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Migrate runs all pending migrations against the database.
func Migrate(ctx context.Context, db *sql.DB) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	currentVersion, err := getCurrentVersion(ctx, db)
	if err != nil {
		return err
	}

	for i := currentVersion; i < len(migrations); i++ {
		if err := runMigration(ctx, db, i+1, migrations[i]); err != nil {
			return fmt.Errorf("migration v%d: %w", i+1, err)
		}
	}
	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	return err
}

func getCurrentVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	err := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	return version, err
}

func runMigration(ctx context.Context, db *sql.DB, version int, ddl string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, ddl); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
		return err
	}
	return tx.Commit()
}
