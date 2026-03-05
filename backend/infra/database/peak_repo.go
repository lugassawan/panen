package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lugassawan/panen/backend/domain/shared"
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
	var hp trailingstop.HoldingPeak
	var updatedAt string
	err := r.db.QueryRowContext(ctx, peakGetByHoldingID, holdingID).Scan(
		&hp.ID, &hp.HoldingID, &hp.PeakPrice, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var parseErr error
	if hp.UpdatedAt, parseErr = parseTime(updatedAt); parseErr != nil {
		return nil, parseErr
	}
	return &hp, nil
}

func (r *PeakRepo) ListByHoldingIDs(
	ctx context.Context,
	holdingIDs []string,
) ([]*trailingstop.HoldingPeak, error) {
	if len(holdingIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(holdingIDs))
	args := make([]any, len(holdingIDs))
	for i, id := range holdingIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT id, holding_id, peak_price, updated_at FROM holding_peaks WHERE holding_id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peaks []*trailingstop.HoldingPeak
	for rows.Next() {
		var hp trailingstop.HoldingPeak
		var updatedAt string
		if err := rows.Scan(&hp.ID, &hp.HoldingID, &hp.PeakPrice, &updatedAt); err != nil {
			return nil, err
		}
		if hp.UpdatedAt, err = parseTime(updatedAt); err != nil {
			return nil, err
		}
		peaks = append(peaks, &hp)
	}
	return peaks, rows.Err()
}
