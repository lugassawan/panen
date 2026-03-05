package presenter

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/usecase"
)

// --- mock registries for screener/watchlist ---

type mockIndexRegistry struct {
	indices map[string][]string
}

func newMockIndexRegistry() *mockIndexRegistry {
	return &mockIndexRegistry{indices: map[string][]string{
		"IDX30":      {"BBCA", "BBRI", "TLKM"},
		"LQ45":       {"BBCA", "BBRI", "TLKM", "ASII"},
		"IDXHIDIV20": {"BBCA", "TLKM"},
	}}
}

func (m *mockIndexRegistry) Tickers(name string) ([]string, bool) {
	t, ok := m.indices[name]
	return t, ok
}

func (m *mockIndexRegistry) Names() []string {
	names := make([]string, 0, len(m.indices))
	for k := range m.indices {
		names = append(names, k)
	}
	return names
}

type mockSectorRegistry struct {
	sectors map[string]string
}

func newMockSectorRegistry() *mockSectorRegistry {
	return &mockSectorRegistry{sectors: map[string]string{
		"BBCA": "Financials",
		"BBRI": "Financials",
		"TLKM": "Communication Services",
		"ASII": "Industrials",
	}}
}

func (m *mockSectorRegistry) SectorOf(ticker string) string {
	return m.sectors[ticker]
}

func (m *mockSectorRegistry) AllSectors() []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range m.sectors {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func newTestScreenerHandler() *ScreenerHandler {
	ctx := context.Background()
	stockRepo := newMockStockRepo()

	// Seed stock data for screening.
	now := time.Now().UTC()
	seeds := []*stock.Data{
		{
			ID: "s1", Ticker: "BBCA", Price: 9000, EPS: 500, BVPS: 4000,
			ROE: 12.5, DER: 0.8, PBV: 2.25, PER: 18,
			DividendYield: 3, PayoutRatio: 40, FetchedAt: now, Source: "mock",
		},
		{
			ID: "s2", Ticker: "BBRI", Price: 5000, EPS: 300, BVPS: 2500,
			ROE: 12.0, DER: 1.5, PBV: 2.0, PER: 16,
			DividendYield: 4, PayoutRatio: 50, FetchedAt: now, Source: "mock",
		},
		{
			ID: "s3", Ticker: "TLKM", Price: 3800, EPS: 200, BVPS: 1500,
			ROE: 13.0, DER: 0.5, PBV: 2.53, PER: 19,
			DividendYield: 5, PayoutRatio: 60, FetchedAt: now, Source: "mock",
		},
	}
	for _, d := range seeds {
		_ = stockRepo.Upsert(ctx, d)
	}

	indexReg := newMockIndexRegistry()
	sectorReg := newMockSectorRegistry()
	svc := usecase.NewScreenerService(stockRepo, indexReg, sectorReg)
	return NewScreenerHandler(ctx, svc)
}

func TestScreenerHandlerRunScreen(t *testing.T) {
	handler := newTestScreenerHandler()

	items, err := handler.RunScreen("INDEX", "IDX30", "MODERATE", "", "score", false, nil)
	if err != nil {
		t.Fatalf("RunScreen() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected screener results")
	}
	for _, item := range items {
		if item.Ticker == "" {
			t.Error("expected non-empty ticker")
		}
	}
}

func TestScreenerHandlerRunScreenSectorFilter(t *testing.T) {
	handler := newTestScreenerHandler()

	items, err := handler.RunScreen("INDEX", "IDX30", "MODERATE", "Financials", "score", false, nil)
	if err != nil {
		t.Fatalf("RunScreen() error = %v", err)
	}
	for _, item := range items {
		if item.Sector != "Financials" {
			t.Errorf("Sector = %q, want %q (filter should apply)", item.Sector, "Financials")
		}
	}
}

func TestScreenerHandlerRunScreenUnknownIndex(t *testing.T) {
	handler := newTestScreenerHandler()

	_, err := handler.RunScreen("INDEX", "NONEXISTENT", "MODERATE", "", "score", false, nil)
	if err == nil {
		t.Error("expected error for unknown index")
	}
}

func TestScreenerHandlerListIndices(t *testing.T) {
	handler := newTestScreenerHandler()

	indices := handler.ListScreenerIndices()
	if len(indices) == 0 {
		t.Error("expected non-empty index list")
	}
}

func TestScreenerHandlerListSectors(t *testing.T) {
	handler := newTestScreenerHandler()

	sectors := handler.ListScreenerSectors()
	if len(sectors) == 0 {
		t.Error("expected non-empty sector list")
	}
}
