package database

import (
	"context"
	"database/sql"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

const (
	snapshotInsert = `INSERT INTO financial_snapshots
		(id, ticker, price, eps, bvps, roe, der, pbv, per, dividend_yield, payout_ratio, source, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	snapshotGetLatest = `SELECT id, ticker, price, eps, bvps, roe, der, pbv, per,
		dividend_yield, payout_ratio, source, fetched_at
		FROM financial_snapshots WHERE ticker = ? AND source = ?
		ORDER BY fetched_at DESC LIMIT 1`
	snapshotCleanup = `DELETE FROM financial_snapshots
		WHERE ticker = ? AND id NOT IN (
			SELECT id FROM financial_snapshots WHERE ticker = ?
			ORDER BY fetched_at DESC LIMIT ?
		)`
)

// SnapshotRepo implements stock.SnapshotRepository.
type SnapshotRepo struct {
	db *sql.DB
}

// NewSnapshotRepo creates a new SnapshotRepo.
func NewSnapshotRepo(db *sql.DB) *SnapshotRepo {
	return &SnapshotRepo{db: db}
}

func (r *SnapshotRepo) Insert(ctx context.Context, data *stock.Data) error {
	_, err := r.db.ExecContext(ctx, snapshotInsert,
		shared.NewID(), data.Ticker, data.Price,
		data.EPS, data.BVPS, data.ROE, data.DER, data.PBV, data.PER,
		data.DividendYield, data.PayoutRatio, data.Source, formatTime(data.FetchedAt))
	return err
}

func (r *SnapshotRepo) GetLatest(ctx context.Context, ticker, source string) (*stock.Data, error) {
	return QueryRow(ctx, r.db, snapshotGetLatest, scanSnapshot, ticker, source)
}

func (r *SnapshotRepo) Cleanup(ctx context.Context, ticker string, keepN int) error {
	_, err := r.db.ExecContext(ctx, snapshotCleanup, ticker, ticker, keepN)
	return err
}

func scanSnapshot(scan func(dest ...any) error) (*stock.Data, error) {
	var d stock.Data
	var fetchedAt string
	if err := scan(
		&d.ID, &d.Ticker, &d.Price,
		&d.EPS, &d.BVPS, &d.ROE, &d.DER, &d.PBV, &d.PER,
		&d.DividendYield, &d.PayoutRatio, &d.Source, &fetchedAt); err != nil {
		return nil, err
	}
	var err error
	if d.FetchedAt, err = parseTime(fetchedAt); err != nil {
		return nil, err
	}
	return &d, nil
}
