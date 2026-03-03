package usecase

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/watchlist"
)

// --- In-memory mock repos ---

type mockWatchlistRepo struct {
	mu    sync.Mutex
	items map[string]*watchlist.Watchlist
}

func newMockWatchlistRepo() *mockWatchlistRepo {
	return &mockWatchlistRepo{items: make(map[string]*watchlist.Watchlist)}
}

func (r *mockWatchlistRepo) Create(_ context.Context, w *watchlist.Watchlist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[w.ID] = w
	return nil
}

func (r *mockWatchlistRepo) GetByID(_ context.Context, id string) (*watchlist.Watchlist, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.items[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return w, nil
}

func (r *mockWatchlistRepo) ListByProfileID(_ context.Context, profileID string) ([]*watchlist.Watchlist, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*watchlist.Watchlist
	for _, w := range r.items {
		if w.ProfileID == profileID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (r *mockWatchlistRepo) Update(_ context.Context, w *watchlist.Watchlist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[w.ID]; !ok {
		return shared.ErrNotFound
	}
	r.items[w.ID] = w
	return nil
}

func (r *mockWatchlistRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return shared.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

type mockWatchlistItemRepo struct {
	mu    sync.Mutex
	items map[string]*watchlist.Item // keyed by watchlistID+":"+ticker
}

func newMockWatchlistItemRepo() *mockWatchlistItemRepo {
	return &mockWatchlistItemRepo{items: make(map[string]*watchlist.Item)}
}

func (r *mockWatchlistItemRepo) Add(_ context.Context, item *watchlist.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := item.WatchlistID + ":" + item.Ticker
	r.items[key] = item
	return nil
}

func (r *mockWatchlistItemRepo) Remove(_ context.Context, watchlistID, ticker string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := watchlistID + ":" + ticker
	if _, ok := r.items[key]; !ok {
		return shared.ErrNotFound
	}
	delete(r.items, key)
	return nil
}

func (r *mockWatchlistItemRepo) ListByWatchlistID(_ context.Context, watchlistID string) ([]*watchlist.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*watchlist.Item
	for _, item := range r.items {
		if item.WatchlistID == watchlistID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *mockWatchlistItemRepo) ExistsByWatchlistAndTicker(
	_ context.Context,
	watchlistID, ticker string,
) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.items[watchlistID+":"+ticker]
	return ok, nil
}

// --- Mock registries ---

type mockIndexRegistry struct {
	data map[string][]string
}

func (r *mockIndexRegistry) Tickers(name string) ([]string, bool) {
	tickers, ok := r.data[name]
	return tickers, ok
}

func (r *mockIndexRegistry) Names() []string {
	names := make([]string, 0, len(r.data))
	for n := range r.data {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

type mockSectorRegistry struct {
	data map[string]string
}

func (r *mockSectorRegistry) SectorOf(ticker string) string {
	return r.data[ticker]
}

func (r *mockSectorRegistry) AllSectors() []string {
	seen := make(map[string]struct{})
	for _, s := range r.data {
		seen[s] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for s := range seen {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

// --- Test setup ---

type watchlistTestFixture struct {
	svc           *WatchlistService
	watchlistRepo *mockWatchlistRepo
	itemRepo      *mockWatchlistItemRepo
	stockRepo     *mockStockRepo
	indexReg      *mockIndexRegistry
	sectorReg     *mockSectorRegistry
	ctx           context.Context
	profileID     string
}

func setupWatchlistTest(t *testing.T) watchlistTestFixture {
	t.Helper()

	watchlistRepo := newMockWatchlistRepo()
	itemRepo := newMockWatchlistItemRepo()
	stockRepo := newMockStockRepo()
	indexReg := &mockIndexRegistry{
		data: map[string][]string{
			"IDX30": {"BBCA", "BBRI", "BMRI"},
			"LQ45":  {"BBCA", "BBRI", "BMRI", "TLKM"},
		},
	}
	sectorReg := &mockSectorRegistry{
		data: map[string]string{
			"BBCA": "Banking",
			"BBRI": "Banking",
			"BMRI": "Banking",
			"TLKM": "Telco",
		},
	}

	svc := NewWatchlistService(watchlistRepo, itemRepo, stockRepo, indexReg, sectorReg)

	return watchlistTestFixture{
		svc:           svc,
		watchlistRepo: watchlistRepo,
		itemRepo:      itemRepo,
		stockRepo:     stockRepo,
		indexReg:      indexReg,
		sectorReg:     sectorReg,
		ctx:           context.Background(),
		profileID:     "profile-1",
	}
}

// --- Tests: CreateWatchlist ---

func TestWatchlistServiceCreateHappy(t *testing.T) {
	f := setupWatchlistTest(t)

	w, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "My Watchlist")
	if err != nil {
		t.Fatalf("CreateWatchlist() error = %v", err)
	}
	if w.ID == "" {
		t.Error("created watchlist should have non-empty ID")
	}
	if w.Name != "My Watchlist" {
		t.Errorf("Name = %q, want %q", w.Name, "My Watchlist")
	}
	if w.ProfileID != f.profileID {
		t.Errorf("ProfileID = %q, want %q", w.ProfileID, f.profileID)
	}
}

func TestWatchlistServiceCreateEmptyName(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "")
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("CreateWatchlist() error = %v, want ErrEmptyName", err)
	}
}

func TestWatchlistServiceCreateWhitespaceName(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "   ")
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("CreateWatchlist() error = %v, want ErrEmptyName", err)
	}
}

func TestWatchlistServiceCreateDuplicateName(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "Favorites")
	if err != nil {
		t.Fatalf("first CreateWatchlist() error = %v", err)
	}

	_, err = f.svc.CreateWatchlist(f.ctx, f.profileID, "Favorites")
	if !errors.Is(err, ErrWatchlistNameTaken) {
		t.Errorf("second CreateWatchlist() error = %v, want ErrWatchlistNameTaken", err)
	}
}

func TestWatchlistServiceCreateDuplicateNameCaseInsensitive(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "favorites")
	if err != nil {
		t.Fatalf("first CreateWatchlist() error = %v", err)
	}

	_, err = f.svc.CreateWatchlist(f.ctx, f.profileID, "FAVORITES")
	if !errors.Is(err, ErrWatchlistNameTaken) {
		t.Errorf("case-insensitive duplicate error = %v, want ErrWatchlistNameTaken", err)
	}
}

func TestWatchlistServiceCreateDifferentProfileSameNameOK(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.CreateWatchlist(f.ctx, "profile-1", "Shared Name")
	if err != nil {
		t.Fatalf("first CreateWatchlist() error = %v", err)
	}

	// Different profile — should succeed.
	_, err = f.svc.CreateWatchlist(f.ctx, "profile-2", "Shared Name")
	if err != nil {
		t.Errorf("CreateWatchlist() for different profile error = %v, want nil", err)
	}
}

// --- Tests: ListWatchlists ---

func TestWatchlistServiceListWatchlistsHappy(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "First")
	if err != nil {
		t.Fatalf("CreateWatchlist() error = %v", err)
	}
	_, err = f.svc.CreateWatchlist(f.ctx, f.profileID, "Second")
	if err != nil {
		t.Fatalf("CreateWatchlist() error = %v", err)
	}

	list, err := f.svc.ListWatchlists(f.ctx, f.profileID)
	if err != nil {
		t.Fatalf("ListWatchlists() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("len(list) = %d, want 2", len(list))
	}
}

func TestWatchlistServiceListWatchlistsEmpty(t *testing.T) {
	f := setupWatchlistTest(t)

	list, err := f.svc.ListWatchlists(f.ctx, "nobody")
	if err != nil {
		t.Fatalf("ListWatchlists() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("len(list) = %d, want 0", len(list))
	}
}

// --- Tests: RenameWatchlist ---

func TestWatchlistServiceRenameHappy(t *testing.T) {
	f := setupWatchlistTest(t)

	w, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "Old Name")
	if err != nil {
		t.Fatalf("CreateWatchlist() error = %v", err)
	}

	if err := f.svc.RenameWatchlist(f.ctx, w.ID, "New Name"); err != nil {
		t.Fatalf("RenameWatchlist() error = %v", err)
	}

	got, _ := f.watchlistRepo.GetByID(f.ctx, w.ID)
	if got.Name != "New Name" {
		t.Errorf("Name = %q, want %q", got.Name, "New Name")
	}
}

func TestWatchlistServiceRenameEmptyName(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Name")

	if err := f.svc.RenameWatchlist(f.ctx, w.ID, ""); !errors.Is(err, ErrEmptyName) {
		t.Errorf("RenameWatchlist() error = %v, want ErrEmptyName", err)
	}
}

func TestWatchlistServiceRenameDuplicateNameTaken(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "Alpha")
	if err != nil {
		t.Fatalf("CreateWatchlist Alpha error = %v", err)
	}
	beta, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "Beta")
	if err != nil {
		t.Fatalf("CreateWatchlist Beta error = %v", err)
	}

	// Try to rename Beta to Alpha — should fail.
	err = f.svc.RenameWatchlist(f.ctx, beta.ID, "Alpha")
	if !errors.Is(err, ErrWatchlistNameTaken) {
		t.Errorf("RenameWatchlist() error = %v, want ErrWatchlistNameTaken", err)
	}
}

func TestWatchlistServiceRenameSameNameOK(t *testing.T) {
	f := setupWatchlistTest(t)

	w, err := f.svc.CreateWatchlist(f.ctx, f.profileID, "Keep It")
	if err != nil {
		t.Fatalf("CreateWatchlist() error = %v", err)
	}

	// Renaming to the same name (self-exclusion) should succeed.
	if err := f.svc.RenameWatchlist(f.ctx, w.ID, "Keep It"); err != nil {
		t.Errorf("RenameWatchlist() same name error = %v, want nil", err)
	}
}

func TestWatchlistServiceRenameNotFound(t *testing.T) {
	f := setupWatchlistTest(t)

	err := f.svc.RenameWatchlist(f.ctx, "nonexistent", "New Name")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("RenameWatchlist() error = %v, want ErrNotFound", err)
	}
}

// --- Tests: DeleteWatchlist ---

func TestWatchlistServiceDeleteHappy(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "To Delete")

	if err := f.svc.DeleteWatchlist(f.ctx, w.ID); err != nil {
		t.Fatalf("DeleteWatchlist() error = %v", err)
	}

	_, err := f.watchlistRepo.GetByID(f.ctx, w.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("watchlist should be deleted, got = %v", err)
	}
}

func TestWatchlistServiceDeleteNotFound(t *testing.T) {
	f := setupWatchlistTest(t)

	err := f.svc.DeleteWatchlist(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("DeleteWatchlist() error = %v, want ErrNotFound", err)
	}
}

// --- Tests: AddTicker / RemoveTicker ---

func TestWatchlistServiceAddTickerHappy(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Tech")

	if err := f.svc.AddTicker(f.ctx, w.ID, "bbca"); err != nil {
		t.Fatalf("AddTicker() error = %v", err)
	}

	// Verify normalized uppercase.
	exists, _ := f.itemRepo.ExistsByWatchlistAndTicker(f.ctx, w.ID, "BBCA")
	if !exists {
		t.Error("BBCA should exist in watchlist after AddTicker with lowercase input")
	}
}

func TestWatchlistServiceAddTickerEmpty(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Tech")

	err := f.svc.AddTicker(f.ctx, w.ID, "")
	if !errors.Is(err, ErrEmptyTicker) {
		t.Errorf("AddTicker() error = %v, want ErrEmptyTicker", err)
	}
}

func TestWatchlistServiceAddTickerDuplicate(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Tech")

	if err := f.svc.AddTicker(f.ctx, w.ID, "BBCA"); err != nil {
		t.Fatalf("first AddTicker() error = %v", err)
	}

	err := f.svc.AddTicker(f.ctx, w.ID, "BBCA")
	if !errors.Is(err, watchlist.ErrTickerAlreadyInWatchlist) {
		t.Errorf("duplicate AddTicker() error = %v, want ErrTickerAlreadyInWatchlist", err)
	}
}

func TestWatchlistServiceRemoveTickerHappy(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Tech")
	_ = f.svc.AddTicker(f.ctx, w.ID, "BBCA")

	if err := f.svc.RemoveTicker(f.ctx, w.ID, "bbca"); err != nil {
		t.Fatalf("RemoveTicker() error = %v", err)
	}

	exists, _ := f.itemRepo.ExistsByWatchlistAndTicker(f.ctx, w.ID, "BBCA")
	if exists {
		t.Error("BBCA should no longer exist after RemoveTicker")
	}
}

func TestWatchlistServiceRemoveTickerEmpty(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Tech")

	err := f.svc.RemoveTicker(f.ctx, w.ID, "")
	if !errors.Is(err, ErrEmptyTicker) {
		t.Errorf("RemoveTicker() error = %v, want ErrEmptyTicker", err)
	}
}

func TestWatchlistServiceRemoveTickerNotFound(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Tech")

	err := f.svc.RemoveTicker(f.ctx, w.ID, "NONEXISTENT")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("RemoveTicker() error = %v, want ErrNotFound", err)
	}
}

// --- Tests: ListItems with enrichment ---

func TestWatchlistServiceListItemsWithStockData(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Banking")
	_ = f.svc.AddTicker(f.ctx, w.ID, "BBCA")
	_ = f.svc.AddTicker(f.ctx, w.ID, "BBRI")

	// Seed stock data for BBCA only.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID:        "sd1",
		Ticker:    "BBCA",
		Price:     8500,
		EPS:       500,
		BVPS:      3000,
		PBV:       2.8,
		PER:       17,
		FetchedAt: time.Now().UTC(),
		Source:    "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	items, err := f.svc.ListItems(f.ctx, w.ID, "")
	if err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	// Find BBCA item.
	var bbca *WatchlistItemWithData
	for _, it := range items {
		if it.Ticker == "BBCA" {
			bbca = it
		}
	}
	if bbca == nil {
		t.Fatal("BBCA not found in items")
	}
	if bbca.StockData == nil {
		t.Error("BBCA StockData should not be nil")
	}
	if bbca.Valuation == nil {
		t.Error("BBCA Valuation should not be nil")
	}
	if bbca.Sector != "Banking" {
		t.Errorf("BBCA Sector = %q, want Banking", bbca.Sector)
	}

	// BBRI has no stock data.
	var bbri *WatchlistItemWithData
	for _, it := range items {
		if it.Ticker == "BBRI" {
			bbri = it
		}
	}
	if bbri == nil {
		t.Fatal("BBRI not found in items")
	}
	if bbri.StockData != nil {
		t.Error("BBRI StockData should be nil (no data seeded)")
	}
	if bbri.Valuation != nil {
		t.Error("BBRI Valuation should be nil (no data seeded)")
	}
}

// --- Tests: sector filter ---

func TestWatchlistServiceListItemsSectorFilter(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Mixed")
	_ = f.svc.AddTicker(f.ctx, w.ID, "BBCA") // Banking
	_ = f.svc.AddTicker(f.ctx, w.ID, "TLKM") // Telco

	items, err := f.svc.ListItems(f.ctx, w.ID, "Banking")
	if err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1 after sector filter", len(items))
	}
	if items[0].Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want BBCA", items[0].Ticker)
	}
}

func TestWatchlistServiceListItemsNoSectorFilter(t *testing.T) {
	f := setupWatchlistTest(t)

	w, _ := f.svc.CreateWatchlist(f.ctx, f.profileID, "Mixed")
	_ = f.svc.AddTicker(f.ctx, w.ID, "BBCA")
	_ = f.svc.AddTicker(f.ctx, w.ID, "TLKM")

	items, err := f.svc.ListItems(f.ctx, w.ID, "")
	if err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2 (no filter)", len(items))
	}
}

// --- Tests: ListPresetItems ---

func TestWatchlistServiceListPresetItemsHappy(t *testing.T) {
	f := setupWatchlistTest(t)

	// Seed some stock data.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 8500,
		EPS: 500, BVPS: 3000, PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	items, err := f.svc.ListPresetItems(f.ctx, "IDX30", "")
	if err != nil {
		t.Fatalf("ListPresetItems() error = %v", err)
	}
	// IDX30 has 3 tickers: BBCA, BBRI, BMRI.
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
}

func TestWatchlistServiceListPresetItemsUnknownIndex(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.svc.ListPresetItems(f.ctx, "NONEXISTENT", "")
	if !errors.Is(err, ErrUnknownIndex) {
		t.Errorf("ListPresetItems() error = %v, want ErrUnknownIndex", err)
	}
}

func TestWatchlistServiceListPresetItemsSectorFilter(t *testing.T) {
	f := setupWatchlistTest(t)

	// LQ45 has BBCA, BBRI, BMRI (Banking) and TLKM (Telco).
	items, err := f.svc.ListPresetItems(f.ctx, "LQ45", "Telco")
	if err != nil {
		t.Fatalf("ListPresetItems() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1 after Telco filter on LQ45", len(items))
	}
	if items[0].Ticker != "TLKM" {
		t.Errorf("Ticker = %q, want TLKM", items[0].Ticker)
	}
}

// --- Tests: ListIndexNames / ListSectors ---

func TestWatchlistServiceListIndexNames(t *testing.T) {
	f := setupWatchlistTest(t)

	names := f.svc.ListIndexNames()
	if len(names) != 2 {
		t.Fatalf("len(names) = %d, want 2", len(names))
	}
	if names[0] != "IDX30" {
		t.Errorf("names[0] = %q, want IDX30", names[0])
	}
	if names[1] != "LQ45" {
		t.Errorf("names[1] = %q, want LQ45", names[1])
	}
}

func TestWatchlistServiceListSectors(t *testing.T) {
	f := setupWatchlistTest(t)

	sectors := f.svc.ListSectors()
	if len(sectors) != 2 {
		t.Fatalf("len(sectors) = %d, want 2", len(sectors))
	}
	// Sorted: Banking, Telco.
	if sectors[0] != "Banking" {
		t.Errorf("sectors[0] = %q, want Banking", sectors[0])
	}
	if sectors[1] != "Telco" {
		t.Errorf("sectors[1] = %q, want Telco", sectors[1])
	}
}
