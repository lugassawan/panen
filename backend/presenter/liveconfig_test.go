package presenter

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/infra/liveconfig"
)

type mockConfigLoader struct {
	reloadCount int
	status      liveconfig.StatusInfo
}

func (m *mockConfigLoader) Reload(_ context.Context) {
	m.reloadCount++
}

func (m *mockConfigLoader) Status() liveconfig.StatusInfo {
	return m.status
}

func newTestHandler() *LiveConfigHandler {
	h := &LiveConfigHandler{}
	h.Init(context.Background())
	return h
}

func TestGetAllConfigStatusEmpty(t *testing.T) {
	h := newTestHandler()
	result := h.GetAllConfigStatus()
	if len(result) != 0 {
		t.Errorf("len = %d, want 0", len(result))
	}
}

func TestGetAllConfigStatus(t *testing.T) {
	h := newTestHandler()
	loader := &mockConfigLoader{
		status: liveconfig.StatusInfo{
			Name:        "test",
			Source:      liveconfig.SourceRemote,
			LastRefresh: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
			Hash:        "abc123",
		},
	}
	h.RegisterLoader("test", loader, nil)

	result := h.GetAllConfigStatus()
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Name != "test" {
		t.Errorf("Name = %q, want test", result[0].Name)
	}
	if result[0].Source != "remote" {
		t.Errorf("Source = %q, want remote", result[0].Source)
	}
	if result[0].DataHash != "abc123" {
		t.Errorf("DataHash = %q, want abc123", result[0].DataHash)
	}
}

func TestForceRefreshAll(t *testing.T) {
	h := newTestHandler()
	loader1 := &mockConfigLoader{}
	loader2 := &mockConfigLoader{}
	reloader1Called := false
	reloader2Called := false

	h.RegisterLoader("a", loader1, func(_ context.Context) { reloader1Called = true })
	h.RegisterLoader("b", loader2, func(_ context.Context) { reloader2Called = true })

	if err := h.ForceRefresh(""); err != nil {
		t.Fatalf("ForceRefresh() error = %v", err)
	}

	if loader1.reloadCount != 1 {
		t.Errorf("loader1 reload count = %d, want 1", loader1.reloadCount)
	}
	if loader2.reloadCount != 1 {
		t.Errorf("loader2 reload count = %d, want 1", loader2.reloadCount)
	}
	if !reloader1Called {
		t.Error("reloader1 not called")
	}
	if !reloader2Called {
		t.Error("reloader2 not called")
	}
}

func TestForceRefreshSpecific(t *testing.T) {
	h := newTestHandler()
	loader := &mockConfigLoader{}
	reloaderCalled := false

	h.RegisterLoader("test", loader, func(_ context.Context) { reloaderCalled = true })

	if err := h.ForceRefresh("test"); err != nil {
		t.Fatalf("ForceRefresh(test) error = %v", err)
	}

	if loader.reloadCount != 1 {
		t.Errorf("reload count = %d, want 1", loader.reloadCount)
	}
	if !reloaderCalled {
		t.Error("reloader not called")
	}
}

func TestForceRefreshUnknown(t *testing.T) {
	h := newTestHandler()
	err := h.ForceRefresh("nonexistent")
	if err == nil {
		t.Error("expected error for unknown config")
	}
}

func TestForceRefreshNilReloader(t *testing.T) {
	h := newTestHandler()
	loader := &mockConfigLoader{}
	h.RegisterLoader("test", loader, nil)

	if err := h.ForceRefresh("test"); err != nil {
		t.Fatalf("ForceRefresh() error = %v", err)
	}
	if loader.reloadCount != 1 {
		t.Errorf("reload count = %d, want 1", loader.reloadCount)
	}
}
