package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// scanFunc converts a row's columns into a value of type T.
// The scan parameter has the same signature as sql.Row.Scan / sql.Rows.Scan.
type scanFunc[T any] func(scan func(dest ...any) error) (T, error)

// queryRow executes a query that returns at most one row and applies scanFn.
// sql.ErrNoRows is mapped to shared.ErrNotFound.
func queryRow[T any](ctx context.Context, db *sql.DB, query string, scanFn scanFunc[T], args ...any) (T, error) {
	row := db.QueryRowContext(ctx, query, args...)
	result, err := scanFn(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		var zero T
		return zero, shared.ErrNotFound
	}
	return result, err
}

// queryAll executes a query that returns multiple rows and applies scanFn to each.
func queryAll[T any](ctx context.Context, db *sql.DB, query string, scanFn scanFunc[T], args ...any) ([]T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		result, err := scanFn(rows.Scan)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}
