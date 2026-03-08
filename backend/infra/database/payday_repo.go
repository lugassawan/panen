package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/payday"
)

const (
	paydayEventInsert = `INSERT INTO payday_events
		(id, month, portfolio_id, expected, actual, status, defer_until, confirmed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	paydayEventGetByMonthAndPortfolio = `SELECT id, month, portfolio_id, expected, actual, status,
		defer_until, confirmed_at, created_at, updated_at
		FROM payday_events WHERE month = ? AND portfolio_id = ?`
	paydayEventListByMonth = `SELECT id, month, portfolio_id, expected, actual, status,
		defer_until, confirmed_at, created_at, updated_at
		FROM payday_events WHERE month = ?`
	paydayEventListByPortfolioID = `SELECT id, month, portfolio_id, expected, actual, status,
		defer_until, confirmed_at, created_at, updated_at
		FROM payday_events WHERE portfolio_id = ?`
	paydayEventUpdate = `UPDATE payday_events SET month = ?, portfolio_id = ?, expected = ?,
		actual = ?, status = ?, defer_until = ?, confirmed_at = ?, updated_at = ?
		WHERE id = ?`
)

// PaydayRepo implements payday.Repository.
type PaydayRepo struct {
	db *sql.DB
}

// NewPaydayRepo creates a new PaydayRepo.
func NewPaydayRepo(db *sql.DB) *PaydayRepo {
	return &PaydayRepo{db: db}
}

func (r *PaydayRepo) Create(ctx context.Context, event *payday.PaydayEvent) error {
	_, err := r.db.ExecContext(ctx, paydayEventInsert,
		event.ID, event.Month, event.PortfolioID, event.Expected, event.Actual,
		string(event.Status), formatNullableTime(event.DeferUntil),
		formatNullableTime(event.ConfirmedAt),
		formatTime(event.CreatedAt), formatTime(event.UpdatedAt))
	return err
}

func (r *PaydayRepo) GetByMonthAndPortfolio(
	ctx context.Context, month, portfolioID string,
) (*payday.PaydayEvent, error) {
	return QueryRow(ctx, r.db, paydayEventGetByMonthAndPortfolio, scanPaydayEvent, month, portfolioID)
}

func (r *PaydayRepo) ListByMonth(ctx context.Context, month string) ([]*payday.PaydayEvent, error) {
	return QueryAll(ctx, r.db, paydayEventListByMonth, scanPaydayEvent, month)
}

func (r *PaydayRepo) ListByPortfolioID(ctx context.Context, portfolioID string) ([]*payday.PaydayEvent, error) {
	return QueryAll(ctx, r.db, paydayEventListByPortfolioID, scanPaydayEvent, portfolioID)
}

func (r *PaydayRepo) Update(ctx context.Context, event *payday.PaydayEvent) error {
	res, err := r.db.ExecContext(ctx, paydayEventUpdate,
		event.Month, event.PortfolioID, event.Expected, event.Actual,
		string(event.Status), formatNullableTime(event.DeferUntil),
		formatNullableTime(event.ConfirmedAt),
		formatTime(event.UpdatedAt), event.ID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func formatNullableTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: formatTime(*t), Valid: true}
}

func scanPaydayEvent(scan func(dest ...any) error) (*payday.PaydayEvent, error) {
	var e payday.PaydayEvent
	var status string
	var deferUntil, confirmedAt sql.NullString
	var createdAt, updatedAt string
	if err := scan(&e.ID, &e.Month, &e.PortfolioID, &e.Expected, &e.Actual,
		&status, &deferUntil, &confirmedAt, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	e.Status = payday.Status(status)
	var err error
	if deferUntil.Valid {
		t, parseErr := parseTime(deferUntil.String)
		if parseErr != nil {
			return nil, parseErr
		}
		e.DeferUntil = &t
	}
	if confirmedAt.Valid {
		t, parseErr := parseTime(confirmedAt.String)
		if parseErr != nil {
			return nil, parseErr
		}
		e.ConfirmedAt = &t
	}
	if e.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if e.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}
