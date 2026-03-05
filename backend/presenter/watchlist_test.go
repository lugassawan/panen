package presenter

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/watchlist"
	"github.com/lugassawan/panen/backend/usecase"
)

// --- mock watchlist repos ---

type mockWatchlistRepo struct {
	watchlists map[string]*watchlist.Watchlist
}

func newMockWatchlistRepo() *mockWatchlistRepo {
	return &mockWatchlistRepo{watchlists: make(map[string]*watchlist.Watchlist)}
}

func (m *mockWatchlistRepo) Create(_ context.Context, w *watchlist.Watchlist) error {
	m.watchlists[w.ID] = w
	return nil
}

func (m *mockWatchlistRepo) GetByID(_ context.Context, id string) (*watchlist.Watchlist, error) {
	w, ok := m.watchlists[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return w, nil
}

func (m *mockWatchlistRepo) ListByProfileID(_ context.Context, profileID string) ([]*watchlist.Watchlist, error) {
	var result []*watchlist.Watchlist
	for _, w := range m.watchlists {
		if w.ProfileID == profileID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *mockWatchlistRepo) Update(_ context.Context, w *watchlist.Watchlist) error {
	m.watchlists[w.ID] = w
	return nil
}

func (m *mockWatchlistRepo) Delete(_ context.Context, id string) error {
	delete(m.watchlists, id)
	return nil
}

type mockWatchlistItemRepo struct {
	items map[string]*watchlist.Item // key: watchlistID+":"+ticker
}

func newMockWatchlistItemRepo() *mockWatchlistItemRepo {
	return &mockWatchlistItemRepo{items: make(map[string]*watchlist.Item)}
}

func (m *mockWatchlistItemRepo) Add(_ context.Context, item *watchlist.Item) error {
	m.items[item.WatchlistID+":"+item.Ticker] = item
	return nil
}

func (m *mockWatchlistItemRepo) Remove(_ context.Context, watchlistID, ticker string) error {
	delete(m.items, watchlistID+":"+ticker)
	return nil
}

func (m *mockWatchlistItemRepo) ListByWatchlistID(_ context.Context, watchlistID string) ([]*watchlist.Item, error) {
	var result []*watchlist.Item
	for _, item := range m.items {
		if item.WatchlistID == watchlistID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockWatchlistItemRepo) ExistsByWatchlistAndTicker(
	_ context.Context,
	watchlistID, ticker string,
) (bool, error) {
	_, ok := m.items[watchlistID+":"+ticker]
	return ok, nil
}

func newTestWatchlistHandler() *WatchlistHandler {
	ctx := context.Background()
	wlRepo := newMockWatchlistRepo()
	itemRepo := newMockWatchlistItemRepo()
	stockRepo := newMockStockRepo()
	indexReg := newMockIndexRegistry()
	sectorReg := newMockSectorRegistry()

	// Seed stock data.
	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID:        "s1",
		Ticker:    "BBCA",
		Price:     9000,
		EPS:       500,
		BVPS:      4000,
		ROE:       12.5,
		FetchedAt: time.Now().UTC(),
		Source:    "mock",
	})

	svc := usecase.NewWatchlistService(wlRepo, itemRepo, stockRepo, indexReg, sectorReg)
	return NewWatchlistHandler(ctx, "profile-1", svc)
}

func TestWatchlistHandlerCreateAndList(t *testing.T) {
	handler := newTestWatchlistHandler()

	w, err := handler.CreateWatchlist("My Watchlist")
	if err != nil {
		t.Fatalf("CreateWatchlist() error = %v", err)
	}
	if w.Name != "My Watchlist" {
		t.Errorf("Name = %q, want %q", w.Name, "My Watchlist")
	}
	if w.ID == "" {
		t.Error("expected non-empty ID")
	}

	list, err := handler.ListWatchlists()
	if err != nil {
		t.Fatalf("ListWatchlists() error = %v", err)
	}
	if len(list) != 1 {
		t.Errorf("got %d watchlists, want 1", len(list))
	}
}

func TestWatchlistHandlerCreateEmptyName(t *testing.T) {
	handler := newTestWatchlistHandler()

	_, err := handler.CreateWatchlist("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestWatchlistHandlerRenameAndDelete(t *testing.T) {
	handler := newTestWatchlistHandler()

	w, _ := handler.CreateWatchlist("Original")
	if err := handler.RenameWatchlist(w.ID, "Renamed"); err != nil {
		t.Fatalf("RenameWatchlist() error = %v", err)
	}

	if err := handler.DeleteWatchlist(w.ID); err != nil {
		t.Fatalf("DeleteWatchlist() error = %v", err)
	}

	list, _ := handler.ListWatchlists()
	if len(list) != 0 {
		t.Errorf("got %d watchlists after delete, want 0", len(list))
	}
}

func TestWatchlistHandlerAddAndRemoveTicker(t *testing.T) {
	handler := newTestWatchlistHandler()

	w, _ := handler.CreateWatchlist("Test WL")
	if err := handler.AddToWatchlist(w.ID, "BBCA"); err != nil {
		t.Fatalf("AddToWatchlist() error = %v", err)
	}

	items, err := handler.GetWatchlistItems(w.ID, "")
	if err != nil {
		t.Fatalf("GetWatchlistItems() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want %q", items[0].Ticker, "BBCA")
	}

	if err := handler.RemoveFromWatchlist(w.ID, "BBCA"); err != nil {
		t.Fatalf("RemoveFromWatchlist() error = %v", err)
	}

	items, _ = handler.GetWatchlistItems(w.ID, "")
	if len(items) != 0 {
		t.Errorf("got %d items after remove, want 0", len(items))
	}
}

func TestWatchlistHandlerListIndexNames(t *testing.T) {
	handler := newTestWatchlistHandler()

	names := handler.ListIndexNames()
	if len(names) == 0 {
		t.Error("expected non-empty index names")
	}
}

func TestWatchlistHandlerListSectors(t *testing.T) {
	handler := newTestWatchlistHandler()

	sectors := handler.ListWatchlistSectors()
	if len(sectors) == 0 {
		t.Error("expected non-empty sectors")
	}
}
