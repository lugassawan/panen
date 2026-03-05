package presenter

import (
	"context"
	"errors"
	"testing"

	"github.com/lugassawan/panen/backend/domain/settings"
)

// mockSettingsRepo implements settings.Repository for testing.
type mockSettingsRepo struct {
	cfg *settings.RefreshSettings
	kv  map[string]string
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{
		cfg: settings.DefaultRefreshSettings(),
		kv:  make(map[string]string),
	}
}

func (r *mockSettingsRepo) GetRefreshSettings(_ context.Context) (*settings.RefreshSettings, error) {
	s := *r.cfg
	return &s, nil
}

func (r *mockSettingsRepo) SaveRefreshSettings(_ context.Context, s *settings.RefreshSettings) error {
	r.cfg = s
	return nil
}

func (r *mockSettingsRepo) GetSetting(_ context.Context, key string) (string, error) {
	return r.kv[key], nil
}

func (r *mockSettingsRepo) SetSetting(_ context.Context, key, value string) error {
	r.kv[key] = value
	return nil
}

func TestRefreshHandlerGetRefreshSettings(t *testing.T) {
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	handler := NewRefreshHandler(ctx, nil, settingsRepo)

	resp, err := handler.GetRefreshSettings()
	if err != nil {
		t.Fatalf("GetRefreshSettings() error = %v", err)
	}
	if !resp.AutoRefreshEnabled {
		t.Error("expected AutoRefreshEnabled = true")
	}
	if resp.IntervalMinutes != 720 {
		t.Errorf("IntervalMinutes = %d, want 720", resp.IntervalMinutes)
	}
}

func TestRefreshHandlerUpdateRefreshSettingsValid(t *testing.T) {
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	handler := NewRefreshHandler(ctx, nil, settingsRepo)

	err := handler.UpdateRefreshSettings(false, 60)
	if err != nil {
		t.Fatalf("UpdateRefreshSettings() error = %v", err)
	}

	resp, _ := handler.GetRefreshSettings()
	if resp.AutoRefreshEnabled {
		t.Error("expected AutoRefreshEnabled = false after update")
	}
	if resp.IntervalMinutes != 60 {
		t.Errorf("IntervalMinutes = %d, want 60", resp.IntervalMinutes)
	}
}

func TestRefreshHandlerUpdateRefreshSettingsInvalidInterval(t *testing.T) {
	ctx := context.Background()
	settingsRepo := newMockSettingsRepo()
	handler := NewRefreshHandler(ctx, nil, settingsRepo)

	tests := []struct {
		name     string
		interval int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.UpdateRefreshSettings(true, tt.interval)
			if !errors.Is(err, errInvalidInterval) {
				t.Errorf("expected errInvalidInterval, got %v", err)
			}
		})
	}
}
