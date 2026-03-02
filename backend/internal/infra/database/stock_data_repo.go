package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/shared"
	"github.com/lugassawan/panen/backend/internal/domain/stock"
)

const (
	stockUpsert = `INSERT INTO stock_data
		(id, ticker, price, high_52_week, low_52_week, eps, bvps, roe, der,
		 pbv, per, dividend_yield, payout_ratio, fetched_at, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ticker, source) DO UPDATE SET
		 id = excluded.id, price = excluded.price,
		 high_52_week = excluded.high_52_week, low_52_week = excluded.low_52_week,
		 eps = excluded.eps, bvps = excluded.bvps, roe = excluded.roe,
		 der = excluded.der, pbv = excluded.pbv, per = excluded.per,
		 dividend_yield = excluded.dividend_yield, payout_ratio = excluded.payout_ratio,
		 fetched_at = excluded.fetched_at`
	stockGetByTicker = `SELECT id, ticker, price, high_52_week, low_52_week,
		eps, bvps, roe, der, pbv, per, dividend_yield, payout_ratio, fetched_at, source
		FROM stock_data WHERE ticker = ? ORDER BY fetched_at DESC LIMIT 1`
	stockGetByTickerAndSource = `SELECT id, ticker, price, high_52_week, low_52_week,
		eps, bvps, roe, der, pbv, per, dividend_yield, payout_ratio, fetched_at, source
		FROM stock_data WHERE ticker = ? AND source = ? ORDER BY fetched_at DESC LIMIT 1`
	stockDeleteOlderThan = `DELETE FROM stock_data WHERE fetched_at < ?`
)

// StockDataRepo implements stock.Repository.
type StockDataRepo struct {
	db *sql.DB
}

// NewStockDataRepo creates a new StockDataRepo.
func NewStockDataRepo(db *sql.DB) *StockDataRepo {
	return &StockDataRepo{db: db}
}

func (r *StockDataRepo) Upsert(ctx context.Context, d *stock.Data) error {
	_, err := r.db.ExecContext(ctx, stockUpsert,
		d.ID, d.Ticker, d.Price, d.High52Week, d.Low52Week,
		d.EPS, d.BVPS, d.ROE, d.DER, d.PBV, d.PER,
		d.DividendYield, d.PayoutRatio, formatTime(d.FetchedAt), d.Source)
	return err
}

func (r *StockDataRepo) GetByTicker(ctx context.Context, ticker string) (*stock.Data, error) {
	return r.scanStockRow(r.db.QueryRowContext(ctx, stockGetByTicker, ticker))
}

func (r *StockDataRepo) GetByTickerAndSource(ctx context.Context, ticker string, source string) (*stock.Data, error) {
	return r.scanStockRow(r.db.QueryRowContext(ctx, stockGetByTickerAndSource, ticker, source))
}

func (r *StockDataRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, stockDeleteOlderThan, formatTime(before))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *StockDataRepo) scanStockRow(row *sql.Row) (*stock.Data, error) {
	var d stock.Data
	var fetchedAt string
	err := row.Scan(
		&d.ID, &d.Ticker, &d.Price, &d.High52Week, &d.Low52Week,
		&d.EPS, &d.BVPS, &d.ROE, &d.DER, &d.PBV, &d.PER,
		&d.DividendYield, &d.PayoutRatio, &fetchedAt, &d.Source)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if d.FetchedAt, err = parseTime(fetchedAt); err != nil {
		return nil, err
	}
	return &d, nil
}
