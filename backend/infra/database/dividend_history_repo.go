package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/shared"
)

const (
	dividendHistoryUpsert = `INSERT INTO dividend_history
		(id, ticker, ex_date, amount, source)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(ticker, ex_date, source) DO UPDATE SET
		amount = excluded.amount`
	dividendHistoryGetByTicker = `SELECT id, ticker, ex_date, amount, source
		FROM dividend_history WHERE ticker = ? AND source = ? ORDER BY ex_date ASC`
	dividendHistoryLatestDate = `SELECT MAX(ex_date) FROM dividend_history WHERE ticker = ? AND source = ?`
)

// DividendHistoryRepo implements dividend.HistoryRepository.
type DividendHistoryRepo struct {
	db *sql.DB
}

// NewDividendHistoryRepo creates a new DividendHistoryRepo.
func NewDividendHistoryRepo(db *sql.DB) *DividendHistoryRepo {
	return &DividendHistoryRepo{db: db}
}

func (r *DividendHistoryRepo) BulkUpsert(ctx context.Context, events []dividend.DividendEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stmt, err := tx.PrepareContext(ctx, dividendHistoryUpsert)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range events {
		e := &events[i]
		if e.ID == "" {
			e.ID = shared.NewID()
		}
		_, err := stmt.ExecContext(ctx,
			e.ID, e.Ticker, formatTime(e.ExDate),
			e.Amount, e.Source)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *DividendHistoryRepo) GetByTicker(
	ctx context.Context,
	ticker, source string,
) ([]dividend.DividendEvent, error) {
	rows, err := r.db.QueryContext(ctx, dividendHistoryGetByTicker, ticker, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []dividend.DividendEvent
	for rows.Next() {
		var e dividend.DividendEvent
		var dateStr string
		if err := rows.Scan(&e.ID, &e.Ticker, &dateStr, &e.Amount, &e.Source); err != nil {
			return nil, err
		}
		if e.ExDate, err = parseTime(dateStr); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *DividendHistoryRepo) LatestDate(
	ctx context.Context,
	ticker, source string,
) (time.Time, error) {
	var dateStr sql.NullString
	err := r.db.QueryRowContext(ctx, dividendHistoryLatestDate, ticker, source).Scan(&dateStr)
	if err != nil {
		return time.Time{}, err
	}
	if !dateStr.Valid {
		return time.Time{}, nil
	}
	return parseTime(dateStr.String)
}
