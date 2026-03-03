package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/usecase"
)

// WatchlistHandler handles watchlist requests.
type WatchlistHandler struct {
	ctx        context.Context
	profileID  string
	watchlists *usecase.WatchlistService
}

// NewWatchlistHandler creates a new WatchlistHandler.
func NewWatchlistHandler(
	ctx context.Context, profileID string, watchlists *usecase.WatchlistService,
) *WatchlistHandler {
	return &WatchlistHandler{ctx: ctx, profileID: profileID, watchlists: watchlists}
}

// ListWatchlists returns all watchlists for the current user.
func (h *WatchlistHandler) ListWatchlists() ([]*WatchlistResponse, error) {
	wls, err := h.watchlists.ListWatchlists(h.ctx, h.profileID)
	if err != nil {
		return nil, err
	}
	result := make([]*WatchlistResponse, len(wls))
	for i, w := range wls {
		result[i] = newWatchlistResponse(w)
	}
	return result, nil
}

// CreateWatchlist creates a new watchlist for the current user.
func (h *WatchlistHandler) CreateWatchlist(name string) (*WatchlistResponse, error) {
	w, err := h.watchlists.CreateWatchlist(h.ctx, h.profileID, name)
	if err != nil {
		return nil, err
	}
	return newWatchlistResponse(w), nil
}

// RenameWatchlist updates the name of the given watchlist.
func (h *WatchlistHandler) RenameWatchlist(id, name string) error {
	return h.watchlists.RenameWatchlist(h.ctx, id, name)
}

// DeleteWatchlist removes the watchlist with the given ID.
func (h *WatchlistHandler) DeleteWatchlist(id string) error {
	return h.watchlists.DeleteWatchlist(h.ctx, id)
}

// AddToWatchlist adds a ticker to the given watchlist.
func (h *WatchlistHandler) AddToWatchlist(watchlistID, ticker string) error {
	return h.watchlists.AddTicker(h.ctx, watchlistID, ticker)
}

// RemoveFromWatchlist removes a ticker from the given watchlist.
func (h *WatchlistHandler) RemoveFromWatchlist(watchlistID, ticker string) error {
	return h.watchlists.RemoveTicker(h.ctx, watchlistID, ticker)
}

// GetWatchlistItems returns all items in the watchlist, enriched with stock data and valuation.
// If sectorFilter is non-empty, only items matching that sector are returned.
func (h *WatchlistHandler) GetWatchlistItems(watchlistID, sectorFilter string) ([]*WatchlistItemResponse, error) {
	items, err := h.watchlists.ListItems(h.ctx, watchlistID, sectorFilter)
	if err != nil {
		return nil, err
	}
	result := make([]*WatchlistItemResponse, len(items))
	for i, item := range items {
		result[i] = newWatchlistItemResponse(item)
	}
	return result, nil
}

// GetPresetItems resolves tickers from the named index and returns enriched items.
func (h *WatchlistHandler) GetPresetItems(indexName, sectorFilter string) ([]*WatchlistItemResponse, error) {
	items, err := h.watchlists.ListPresetItems(h.ctx, indexName, sectorFilter)
	if err != nil {
		return nil, err
	}
	result := make([]*WatchlistItemResponse, len(items))
	for i, item := range items {
		result[i] = newWatchlistItemResponse(item)
	}
	return result, nil
}

// ListIndexNames returns all registered index names.
func (h *WatchlistHandler) ListIndexNames() []string {
	return h.watchlists.ListIndexNames()
}

// ListWatchlistSectors returns all unique sector names.
func (h *WatchlistHandler) ListWatchlistSectors() []string {
	return h.watchlists.ListSectors()
}
