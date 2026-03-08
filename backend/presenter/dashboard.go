package presenter

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/usecase"
)

// DashboardHandler handles dashboard requests from the frontend.
type DashboardHandler struct {
	ctx       context.Context
	dashboard *usecase.DashboardService
}

// Bind injects dependencies after construction.
func (h *DashboardHandler) Bind(ctx context.Context, dashboard *usecase.DashboardService) {
	h.ctx = ctx
	h.dashboard = dashboard
}

// GetDashboardOverview returns aggregated performance data across all portfolios.
func (h *DashboardHandler) GetDashboardOverview() (*DashboardOverviewResponse, error) {
	if h.dashboard == nil {
		return emptyDashboardOverview(), nil
	}

	overview, err := h.dashboard.GetOverview(h.ctx)
	if err != nil {
		return nil, fmt.Errorf("get dashboard overview: %w", err)
	}

	return newDashboardOverviewResponse(overview), nil
}
