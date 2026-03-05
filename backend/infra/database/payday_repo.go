package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/shared"
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
	var e payday.PaydayEvent
	var status string
	var deferUntil, confirmedAt sql.NullString
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, paydayEventGetByMonthAndPortfolio, month, portfolioID).Scan(
		&e.ID, &e.Month, &e.PortfolioID, &e.Expected, &e.Actual, &status,
		&deferUntil, &confirmedAt, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return scanPaydayEvent(&e, status, deferUntil, confirmedAt, createdAt, updatedAt)
}

func (r *PaydayRepo) ListByMonth(ctx context.Context, month string) ([]*payday.PaydayEvent, error) {
	rows, err := r.db.QueryContext(ctx, paydayEventListByMonth, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPaydayEvents(rows)
}

func (r *PaydayRepo) ListByPortfolioID(ctx context.Context, portfolioID string) ([]*payday.PaydayEvent, error) {
	rows, err := r.db.QueryContext(ctx, paydayEventListByPortfolioID, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPaydayEvents(rows)
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

func scanPaydayEvents(rows *sql.Rows) ([]*payday.PaydayEvent, error) {
	var events []*payday.PaydayEvent
	for rows.Next() {
		var e payday.PaydayEvent
		var status string
		var deferUntil, confirmedAt sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&e.ID, &e.Month, &e.PortfolioID, &e.Expected, &e.Actual,
			&status, &deferUntil, &confirmedAt, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		ev, err := scanPaydayEvent(&e, status, deferUntil, confirmedAt, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func scanPaydayEvent(
	e *payday.PaydayEvent,
	status string,
	deferUntil, confirmedAt sql.NullString,
	createdAt, updatedAt string,
) (*payday.PaydayEvent, error) {
	e.Status = payday.Status(status)
	var err error
	if deferUntil.Valid {
		t, err := parseTime(deferUntil.String)
		if err != nil {
			return nil, err
		}
		e.DeferUntil = &t
	}
	if confirmedAt.Valid {
		t, err := parseTime(confirmedAt.String)
		if err != nil {
			return nil, err
		}
		e.ConfirmedAt = &t
	}
	if e.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if e.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return e, nil
}
