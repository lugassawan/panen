package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

const (
	priceHistoryUpsert = `INSERT INTO price_history
		(id, ticker, date, open, high, low, close, volume, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ticker, date, source) DO UPDATE SET
		open = excluded.open, high = excluded.high,
		low = excluded.low, close = excluded.close, volume = excluded.volume`
	priceHistoryGetByTicker = `SELECT id, ticker, date, open, high, low, close, volume, source
		FROM price_history WHERE ticker = ? AND source = ? ORDER BY date ASC`
	priceHistoryLatestDate     = `SELECT MAX(date) FROM price_history WHERE ticker = ? AND source = ?`
	priceHistoryDeleteByTicker = `DELETE FROM price_history WHERE ticker = ?`
)

// PriceHistoryRepo implements stock.PriceHistoryRepository.
type PriceHistoryRepo struct {
	db *sql.DB
}

// NewPriceHistoryRepo creates a new PriceHistoryRepo.
func NewPriceHistoryRepo(db *sql.DB) *PriceHistoryRepo {
	return &PriceHistoryRepo{db: db}
}

func (r *PriceHistoryRepo) BulkUpsert(ctx context.Context, points []stock.PricePoint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stmt, err := tx.PrepareContext(ctx, priceHistoryUpsert)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range points {
		p := &points[i]
		if p.ID == "" {
			p.ID = shared.NewID()
		}
		_, err := stmt.ExecContext(ctx,
			p.ID, p.Ticker, formatTime(p.Date),
			p.Open, p.High, p.Low, p.Close, p.Volume, p.Source)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PriceHistoryRepo) GetByTicker(
	ctx context.Context,
	ticker, source string,
) ([]stock.PricePoint, error) {
	return queryAll(ctx, r.db, priceHistoryGetByTicker, scanPricePoint, ticker, source)
}

func (r *PriceHistoryRepo) LatestDate(
	ctx context.Context,
	ticker, source string,
) (time.Time, error) {
	var dateStr sql.NullString
	err := r.db.QueryRowContext(ctx, priceHistoryLatestDate, ticker, source).Scan(&dateStr)
	if err != nil {
		return time.Time{}, err
	}
	if !dateStr.Valid {
		return time.Time{}, nil
	}
	return parseTime(dateStr.String)
}

func (r *PriceHistoryRepo) DeleteByTicker(ctx context.Context, ticker string) error {
	_, err := r.db.ExecContext(ctx, priceHistoryDeleteByTicker, ticker)
	return err
}

func scanPricePoint(scan func(dest ...any) error) (stock.PricePoint, error) {
	var pp stock.PricePoint
	var dateStr string
	if err := scan(
		&pp.ID, &pp.Ticker, &dateStr,
		&pp.Open, &pp.High, &pp.Low, &pp.Close, &pp.Volume, &pp.Source,
	); err != nil {
		return pp, err
	}
	var err error
	if pp.Date, err = parseTime(dateStr); err != nil {
		return pp, err
	}
	return pp, nil
}
