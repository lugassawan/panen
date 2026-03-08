package database

import (
	"context"
	"database/sql"

	"github.com/lugassawan/panen/backend/domain/trailingstop"
)

const (
	peakUpsert = `INSERT INTO holding_peaks
		(id, holding_id, peak_price, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(holding_id) DO UPDATE SET
		peak_price = excluded.peak_price, updated_at = excluded.updated_at`
	peakGetByHoldingID = `SELECT id, holding_id, peak_price, updated_at
		FROM holding_peaks WHERE holding_id = ?`
)

// PeakRepo implements trailingstop.PeakRepository.
type PeakRepo struct {
	db *sql.DB
}

// NewPeakRepo creates a new PeakRepo.
func NewPeakRepo(db *sql.DB) *PeakRepo {
	return &PeakRepo{db: db}
}

func (r *PeakRepo) Upsert(ctx context.Context, peak *trailingstop.HoldingPeak) error {
	_, err := r.db.ExecContext(ctx, peakUpsert,
		peak.ID, peak.HoldingID, peak.PeakPrice, formatTime(peak.UpdatedAt))
	return err
}

func (r *PeakRepo) GetByHoldingID(
	ctx context.Context,
	holdingID string,
) (*trailingstop.HoldingPeak, error) {
	return QueryRow(ctx, r.db, peakGetByHoldingID, scanPeak, holdingID)
}

func (r *PeakRepo) ListByHoldingIDs(
	ctx context.Context,
	holdingIDs []string,
) ([]*trailingstop.HoldingPeak, error) {
	if len(holdingIDs) == 0 {
		return nil, nil
	}

	args := make([]any, len(holdingIDs))
	for i, id := range holdingIDs {
		args[i] = id
	}

	query := buildINQuery(
		`SELECT id, holding_id, peak_price, updated_at FROM holding_peaks WHERE holding_id IN`,
		len(holdingIDs),
	)

	return QueryAll(ctx, r.db, query, scanPeak, args...)
}

func scanPeak(scan func(dest ...any) error) (*trailingstop.HoldingPeak, error) {
	var hp trailingstop.HoldingPeak
	var updatedAt string
	if err := scan(&hp.ID, &hp.HoldingID, &hp.PeakPrice, &updatedAt); err != nil {
		return nil, err
	}
	var err error
	if hp.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &hp, nil
}
