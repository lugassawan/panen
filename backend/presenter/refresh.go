package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/usecase"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsEmitter wraps the Wails runtime event emitter to satisfy usecase.EventEmitter.
type WailsEmitter struct {
	ctx context.Context
}

// NewWailsEmitter creates a new WailsEmitter bound to the given Wails context.
func NewWailsEmitter(ctx context.Context) *WailsEmitter {
	return &WailsEmitter{ctx: ctx}
}

// Emit sends a named event with data to the frontend via Wails runtime.
func (e *WailsEmitter) Emit(eventName string, data any) {
	runtime.EventsEmit(e.ctx, eventName, data)
}

// RefreshHandler handles refresh-related requests from the frontend.
type RefreshHandler struct {
	ctx      context.Context
	refresh  *usecase.RefreshService
	settings settings.Repository
}

// NewRefreshHandler creates a new RefreshHandler.
func NewRefreshHandler(
	ctx context.Context,
	refresh *usecase.RefreshService,
	settings settings.Repository,
) *RefreshHandler {
	return &RefreshHandler{ctx: ctx, refresh: refresh, settings: settings}
}

// TriggerRefresh triggers an immediate refresh of all tracked stock data.
func (h *RefreshHandler) TriggerRefresh() error {
	return h.refresh.RunNow(h.ctx)
}

// GetRefreshStatus returns the current refresh status.
func (h *RefreshHandler) GetRefreshStatus() *RefreshStatusResponse {
	s := h.refresh.GetStatus()
	return &RefreshStatusResponse{
		State:       s.State,
		LastRefresh: s.LastRefresh,
		Error:       s.Error,
	}
}

// GetRefreshSettings returns the current refresh settings.
func (h *RefreshHandler) GetRefreshSettings() (*RefreshSettingsResponse, error) {
	cfg, err := h.settings.GetRefreshSettings(h.ctx)
	if err != nil {
		return nil, err
	}
	return &RefreshSettingsResponse{
		AutoRefreshEnabled: cfg.AutoRefreshEnabled,
		IntervalMinutes:    cfg.IntervalMinutes,
		LastRefreshedAt:    formatDTO(cfg.LastRefreshedAt),
	}, nil
}

// UpdateRefreshSettings updates the auto-refresh enabled flag and interval.
func (h *RefreshHandler) UpdateRefreshSettings(enabled bool, intervalMinutes int) error {
	cfg, err := h.settings.GetRefreshSettings(h.ctx)
	if err != nil {
		return err
	}
	cfg.AutoRefreshEnabled = enabled
	cfg.IntervalMinutes = intervalMinutes
	return h.settings.SaveRefreshSettings(h.ctx, cfg)
}
