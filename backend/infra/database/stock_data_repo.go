package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
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
	stockListAllTickers  = `SELECT DISTINCT ticker FROM stock_data ORDER BY ticker`
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
	return queryRow(ctx, r.db, stockGetByTicker, scanStockData, ticker)
}

func (r *StockDataRepo) GetByTickerAndSource(ctx context.Context, ticker string, source string) (*stock.Data, error) {
	return queryRow(ctx, r.db, stockGetByTickerAndSource, scanStockData, ticker, source)
}

func (r *StockDataRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, stockDeleteOlderThan, formatTime(before))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *StockDataRepo) ListAllTickers(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, stockListAllTickers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tickers []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tickers = append(tickers, t)
	}
	return tickers, rows.Err()
}

func scanStockData(scan func(dest ...any) error) (*stock.Data, error) {
	var d stock.Data
	var fetchedAt string
	if err := scan(
		&d.ID, &d.Ticker, &d.Price, &d.High52Week, &d.Low52Week,
		&d.EPS, &d.BVPS, &d.ROE, &d.DER, &d.PBV, &d.PER,
		&d.DividendYield, &d.PayoutRatio, &fetchedAt, &d.Source); err != nil {
		return nil, err
	}
	var err error
	if d.FetchedAt, err = parseTime(fetchedAt); err != nil {
		return nil, err
	}
	return &d, nil
}
