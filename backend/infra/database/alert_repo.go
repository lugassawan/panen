package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/alert"
)

const (
	alertInsert = `INSERT INTO fundamental_alerts
		(id, ticker, metric, severity, old_value, new_value, change_pct, status, detected_at, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	alertGetByTicker = `SELECT id, ticker, metric, severity, old_value, new_value, change_pct,
		status, detected_at, resolved_at
		FROM fundamental_alerts WHERE ticker = ? ORDER BY detected_at DESC`
	alertGetActive = `SELECT id, ticker, metric, severity, old_value, new_value, change_pct,
		status, detected_at, resolved_at
		FROM fundamental_alerts WHERE status = 'ACTIVE' ORDER BY detected_at DESC`
	alertGetActiveByTicker = `SELECT id, ticker, metric, severity, old_value, new_value, change_pct,
		status, detected_at, resolved_at
		FROM fundamental_alerts WHERE ticker = ? AND status = 'ACTIVE' ORDER BY detected_at DESC`
	alertAcknowledge = `UPDATE fundamental_alerts SET status = 'ACKNOWLEDGED' WHERE id = ? AND status = 'ACTIVE'`
	alertResolve     = `UPDATE fundamental_alerts SET status = 'RESOLVED', resolved_at = ?
		WHERE id = ? AND status IN ('ACTIVE', 'ACKNOWLEDGED')`
	alertCountActive    = `SELECT COUNT(*) FROM fundamental_alerts WHERE status = 'ACTIVE'`
	alertDeleteOlderThn = `DELETE FROM fundamental_alerts WHERE detected_at < ?`
)

// AlertRepo implements alert.Repository.
type AlertRepo struct {
	db *sql.DB
}

// NewAlertRepo creates a new AlertRepo.
func NewAlertRepo(db *sql.DB) *AlertRepo {
	return &AlertRepo{db: db}
}

func (r *AlertRepo) Create(ctx context.Context, a *alert.FundamentalAlert) error {
	var resolvedAt *string
	if a.ResolvedAt != nil {
		s := formatTime(*a.ResolvedAt)
		resolvedAt = &s
	}
	_, err := r.db.ExecContext(ctx, alertInsert,
		a.ID, a.Ticker, a.Metric, string(a.Severity),
		a.OldValue, a.NewValue, a.ChangePct,
		string(a.Status), formatTime(a.DetectedAt), resolvedAt)
	return err
}

func (r *AlertRepo) GetByTicker(ctx context.Context, ticker string) ([]*alert.FundamentalAlert, error) {
	return r.scanAlerts(r.db.QueryContext(ctx, alertGetByTicker, ticker))
}

func (r *AlertRepo) GetActive(ctx context.Context) ([]*alert.FundamentalAlert, error) {
	return r.scanAlerts(r.db.QueryContext(ctx, alertGetActive))
}

func (r *AlertRepo) GetActiveByTicker(ctx context.Context, ticker string) ([]*alert.FundamentalAlert, error) {
	return r.scanAlerts(r.db.QueryContext(ctx, alertGetActiveByTicker, ticker))
}

func (r *AlertRepo) Acknowledge(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, alertAcknowledge, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *AlertRepo) Resolve(ctx context.Context, id string) error {
	now := formatTime(time.Now().UTC())
	res, err := r.db.ExecContext(ctx, alertResolve, now, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *AlertRepo) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, alertCountActive).Scan(&count)
	return count, err
}

func (r *AlertRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, alertDeleteOlderThn, formatTime(before))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *AlertRepo) scanAlerts(rows *sql.Rows, err error) ([]*alert.FundamentalAlert, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*alert.FundamentalAlert
	for rows.Next() {
		a, scanErr := r.scanAlertRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

func (r *AlertRepo) scanAlertRow(rows *sql.Rows) (*alert.FundamentalAlert, error) {
	var a alert.FundamentalAlert
	var severity, status, detectedAt string
	var resolvedAt sql.NullString

	if err := rows.Scan(
		&a.ID, &a.Ticker, &a.Metric, &severity,
		&a.OldValue, &a.NewValue, &a.ChangePct,
		&status, &detectedAt, &resolvedAt,
	); err != nil {
		return nil, err
	}

	a.Severity = alert.Severity(severity)
	a.Status = alert.AlertStatus(status)

	var err error
	if a.DetectedAt, err = parseTime(detectedAt); err != nil {
		return nil, err
	}
	if resolvedAt.Valid {
		t, parseErr := parseTime(resolvedAt.String)
		if parseErr != nil {
			return nil, parseErr
		}
		a.ResolvedAt = &t
	}
	return &a, nil
}
