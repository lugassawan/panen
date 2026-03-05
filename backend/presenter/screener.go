package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/usecase"
)

// ScreenerHandler handles screener requests.
type ScreenerHandler struct {
	ctx      context.Context
	screener *usecase.ScreenerService
}

// NewScreenerHandler creates a new ScreenerHandler.
func NewScreenerHandler(ctx context.Context, screener *usecase.ScreenerService) *ScreenerHandler {
	return &ScreenerHandler{ctx: ctx, screener: screener}
}

// RunScreen executes a stock screen and returns results.
func (h *ScreenerHandler) RunScreen(
	universeType, universeName, riskProfile, sectorFilter, sortField string,
	sortAsc bool,
	customTickers []string,
) ([]*ScreenerItemResponse, error) {
	results, err := h.screener.Screen(h.ctx, usecase.ScreenRequest{
		UniverseType:  usecase.UniverseType(universeType),
		UniverseName:  universeName,
		CustomTickers: customTickers,
		RiskProfile:   riskProfile,
		SectorFilter:  sectorFilter,
		SortField:     sortField,
		SortAsc:       sortAsc,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*ScreenerItemResponse, len(results))
	for i, r := range results {
		items[i] = newScreenerItemResponse(r)
	}
	return items, nil
}

// ListScreenerIndices returns all registered index names.
func (h *ScreenerHandler) ListScreenerIndices() []string {
	return h.screener.ListIndexNames()
}

// ListScreenerSectors returns all unique sector names.
func (h *ScreenerHandler) ListScreenerSectors() []string {
	return h.screener.ListSectors()
}
