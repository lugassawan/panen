package database

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/settings"
)

func TestSettingsRepoDefaultsAfterMigration(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	repo := NewSettingsRepo(conn)

	got, err := repo.GetRefreshSettings(ctx)
	if err != nil {
		t.Fatalf("GetRefreshSettings() error = %v", err)
	}
	if !got.AutoRefreshEnabled {
		t.Error("AutoRefreshEnabled = false, want true")
	}
	if got.IntervalMinutes != 720 {
		t.Errorf("IntervalMinutes = %d, want 720", got.IntervalMinutes)
	}
	if !got.LastRefreshedAt.IsZero() {
		t.Errorf("LastRefreshedAt = %v, want zero time", got.LastRefreshedAt)
	}
}

func TestSettingsRepoSaveAndGetRoundTrip(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	repo := NewSettingsRepo(conn)

	lastRefreshed := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	want := &settings.RefreshSettings{
		AutoRefreshEnabled: false,
		IntervalMinutes:    60,
		LastRefreshedAt:    lastRefreshed,
	}
	if err := repo.SaveRefreshSettings(ctx, want); err != nil {
		t.Fatalf("SaveRefreshSettings() error = %v", err)
	}

	got, err := repo.GetRefreshSettings(ctx)
	if err != nil {
		t.Fatalf("GetRefreshSettings() error = %v", err)
	}
	if got.AutoRefreshEnabled != want.AutoRefreshEnabled {
		t.Errorf("AutoRefreshEnabled = %v, want %v", got.AutoRefreshEnabled, want.AutoRefreshEnabled)
	}
	if got.IntervalMinutes != want.IntervalMinutes {
		t.Errorf("IntervalMinutes = %d, want %d", got.IntervalMinutes, want.IntervalMinutes)
	}
	if !got.LastRefreshedAt.Equal(want.LastRefreshedAt) {
		t.Errorf("LastRefreshedAt = %v, want %v", got.LastRefreshedAt, want.LastRefreshedAt)
	}
}

func TestSettingsRepoPartialUpdate(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	repo := NewSettingsRepo(conn)

	// First, save with non-default values
	lastRefreshed := time.Date(2024, 3, 1, 8, 0, 0, 0, time.UTC)
	initial := &settings.RefreshSettings{
		AutoRefreshEnabled: false,
		IntervalMinutes:    30,
		LastRefreshedAt:    lastRefreshed,
	}
	if err := repo.SaveRefreshSettings(ctx, initial); err != nil {
		t.Fatalf("SaveRefreshSettings() initial error = %v", err)
	}

	// Now save again changing only the interval
	updated := &settings.RefreshSettings{
		AutoRefreshEnabled: false,
		IntervalMinutes:    1440,
		LastRefreshedAt:    lastRefreshed,
	}
	if err := repo.SaveRefreshSettings(ctx, updated); err != nil {
		t.Fatalf("SaveRefreshSettings() updated error = %v", err)
	}

	got, err := repo.GetRefreshSettings(ctx)
	if err != nil {
		t.Fatalf("GetRefreshSettings() error = %v", err)
	}
	if got.AutoRefreshEnabled != false {
		t.Error("AutoRefreshEnabled changed unexpectedly")
	}
	if got.IntervalMinutes != 1440 {
		t.Errorf("IntervalMinutes = %d, want 1440", got.IntervalMinutes)
	}
	if !got.LastRefreshedAt.Equal(lastRefreshed) {
		t.Errorf("LastRefreshedAt = %v, want %v", got.LastRefreshedAt, lastRefreshed)
	}
}

func TestSettingsRepoSaveZeroTimeClearsLastRefreshed(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	repo := NewSettingsRepo(conn)

	// Save with a real timestamp
	withTime := &settings.RefreshSettings{
		AutoRefreshEnabled: true,
		IntervalMinutes:    720,
		LastRefreshedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.SaveRefreshSettings(ctx, withTime); err != nil {
		t.Fatalf("SaveRefreshSettings() error = %v", err)
	}

	// Save again with zero time to clear it
	cleared := &settings.RefreshSettings{
		AutoRefreshEnabled: true,
		IntervalMinutes:    720,
	}
	if err := repo.SaveRefreshSettings(ctx, cleared); err != nil {
		t.Fatalf("SaveRefreshSettings() clear error = %v", err)
	}

	got, err := repo.GetRefreshSettings(ctx)
	if err != nil {
		t.Fatalf("GetRefreshSettings() error = %v", err)
	}
	if !got.LastRefreshedAt.IsZero() {
		t.Errorf("LastRefreshedAt = %v, want zero time", got.LastRefreshedAt)
	}
}
