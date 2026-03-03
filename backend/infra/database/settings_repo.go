package database

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/lugassawan/panen/backend/domain/settings"
)

const (
	settingsGet = `SELECT key, value FROM app_settings
		WHERE key IN ('auto_refresh_enabled', 'refresh_interval_minutes', 'last_refreshed_at')`
	settingsUpsert = `INSERT INTO app_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`
)

// SettingsRepo implements settings.Repository.
type SettingsRepo struct {
	db *sql.DB
}

// NewSettingsRepo creates a new SettingsRepo.
func NewSettingsRepo(db *sql.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) GetRefreshSettings(ctx context.Context) (*settings.RefreshSettings, error) {
	rows, err := r.db.QueryContext(ctx, settingsGet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s := settings.DefaultRefreshSettings()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		if err := applySettingsKey(s, key, value); err != nil {
			return nil, err
		}
	}
	return s, rows.Err()
}

func applySettingsKey(s *settings.RefreshSettings, key, value string) error {
	switch key {
	case "auto_refresh_enabled":
		s.AutoRefreshEnabled = value == "1"
	case "refresh_interval_minutes":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		s.IntervalMinutes = n
	case "last_refreshed_at":
		if value == "" {
			return nil
		}
		t, err := parseTime(value)
		if err != nil {
			return err
		}
		s.LastRefreshedAt = t
	}
	return nil
}

func (r *SettingsRepo) SaveRefreshSettings(ctx context.Context, s *settings.RefreshSettings) error {
	autoRefresh := "0"
	if s.AutoRefreshEnabled {
		autoRefresh = "1"
	}
	interval := strconv.Itoa(s.IntervalMinutes)

	var lastRefreshed string
	if !s.LastRefreshedAt.IsZero() {
		lastRefreshed = formatTime(s.LastRefreshedAt)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	pairs := [][2]string{
		{"auto_refresh_enabled", autoRefresh},
		{"refresh_interval_minutes", interval},
		{"last_refreshed_at", lastRefreshed},
	}
	for _, p := range pairs {
		if _, err := tx.ExecContext(ctx, settingsUpsert, p[0], p[1]); err != nil {
			return err
		}
	}
	return tx.Commit()
}
