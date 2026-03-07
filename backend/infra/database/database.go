package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps a sql.DB connection to a SQLite database.
type DB struct {
	conn *sql.DB
}

// Open creates a new database connection and applies required pragmas.
func Open(dsn string) (*DB, error) {
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := applyPragmas(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("apply pragmas: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying *sql.DB for use by repositories.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Checkpoint flushes the WAL into the main database file and truncates the WAL.
// This must be called before copying the database file to ensure a consistent backup.
func (db *DB) Checkpoint() error {
	_, err := db.conn.ExecContext(context.Background(), "PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// QuickCheck runs a lightweight integrity check on the database.
// Returns nil if the database is healthy, or an error describing the corruption.
func (db *DB) QuickCheck() error {
	var result string
	if err := db.conn.QueryRowContext(context.Background(), "PRAGMA quick_check").Scan(&result); err != nil {
		return fmt.Errorf("quick_check query: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("quick_check failed: %s", result)
	}
	return nil
}

func applyPragmas(conn *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := conn.ExecContext(context.Background(), p); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}
