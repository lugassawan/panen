package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

func newScreenerTestEnv() (*ScreenerService, *mockStockRepo) {
	stockRepo := newMockStockRepo()
	indexReg := &mockIndexRegistry{
		data: map[string][]string{
			"IDX30": {"BBCA", "BMRI", "TLKM"},
		},
	}
	sectorReg := &mockSectorRegistry{
		data: map[string]string{
			"BBCA": "Banking",
			"BMRI": "Banking",
			"TLKM": "Telecom",
			"ASII": "Automotive",
		},
	}
	svc := NewScreenerService(stockRepo, indexReg, sectorReg)
	return svc, stockRepo
}

func seedStock(t *testing.T, repo *mockStockRepo, d *stock.Data) {
	t.Helper()
	if err := repo.Upsert(context.Background(), d); err != nil {
		t.Fatalf("seedStock %s: %v", d.Ticker, err)
	}
}

func newStock(ticker string, roe, der, price, eps, bvps float64) *stock.Data {
	return &stock.Data{
		Ticker:    ticker,
		ROE:       roe,
		DER:       der,
		Price:     price,
		EPS:       eps,
		BVPS:      bvps,
		PBV:       price / bvps,
		PER:       price / eps,
		FetchedAt: time.Now(),
	}
}

func TestScreenByIndex(t *testing.T) {
	svc, repo := newScreenerTestEnv()
	seedStock(t, repo, newStock("BBCA", 20, 0.5, 9000, 500, 3000))
	seedStock(t, repo, newStock("BMRI", 15, 0.9, 6000, 400, 2500))
	seedStock(t, repo, newStock("TLKM", 10, 0.6, 3000, 200, 1500))

	results, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType: UniverseIndex,
		UniverseName: "IDX30",
		RiskProfile:  "CONSERVATIVE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}

	// Default sort is by score desc
	if results[0].Score < results[1].Score {
		t.Error("results not sorted by score descending")
	}
}

func TestScreenBySector(t *testing.T) {
	svc, repo := newScreenerTestEnv()
	seedStock(t, repo, newStock("BBCA", 20, 0.5, 9000, 500, 3000))
	seedStock(t, repo, newStock("BMRI", 15, 0.9, 6000, 400, 2500))
	seedStock(t, repo, newStock("TLKM", 10, 0.6, 3000, 200, 1500))
	seedStock(t, repo, newStock("ASII", 12, 0.7, 5000, 300, 2000))

	results, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType: UniverseSector,
		UniverseName: "Banking",
		RiskProfile:  "MODERATE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.Sector != "Banking" {
			t.Errorf("unexpected sector %s", r.Sector)
		}
	}
}

func TestScreenCustomTickers(t *testing.T) {
	svc, repo := newScreenerTestEnv()
	seedStock(t, repo, newStock("BBCA", 20, 0.5, 9000, 500, 3000))
	seedStock(t, repo, newStock("TLKM", 10, 0.6, 3000, 200, 1500))

	results, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType:  UniverseCustom,
		CustomTickers: []string{"BBCA", "TLKM"},
		RiskProfile:   "AGGRESSIVE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestScreenWithSectorFilter(t *testing.T) {
	svc, repo := newScreenerTestEnv()
	seedStock(t, repo, newStock("BBCA", 20, 0.5, 9000, 500, 3000))
	seedStock(t, repo, newStock("BMRI", 15, 0.9, 6000, 400, 2500))
	seedStock(t, repo, newStock("TLKM", 10, 0.6, 3000, 200, 1500))

	results, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType: UniverseIndex,
		UniverseName: "IDX30",
		RiskProfile:  "MODERATE",
		SectorFilter: "Banking",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2 (Banking only)", len(results))
	}
}

func TestScreenSorting(t *testing.T) {
	svc, repo := newScreenerTestEnv()
	seedStock(t, repo, newStock("BBCA", 20, 0.5, 9000, 500, 3000))
	seedStock(t, repo, newStock("BMRI", 15, 0.9, 6000, 400, 2500))

	results, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType: UniverseIndex,
		UniverseName: "IDX30",
		RiskProfile:  "MODERATE",
		SortField:    "price",
		SortAsc:      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// TLKM has no data, so it goes last (value = -1e18) when ascending... but actually
	// it has no seedStock, so StockData is nil → price = -1e18
	// With asc: TLKM (-1e18), BMRI (6000), BBCA (9000)
	if len(results) < 2 {
		t.Fatalf("got %d results, want at least 2", len(results))
	}
	// Last two with data should be sorted ascending by price
	dataResults := make([]*ScreenResult, 0)
	for _, r := range results {
		if r.StockData != nil {
			dataResults = append(dataResults, r)
		}
	}
	if len(dataResults) == 2 && dataResults[0].StockData.Price > dataResults[1].StockData.Price {
		t.Error("results not sorted by price ascending")
	}
}

func TestScreenEmptyUniverse(t *testing.T) {
	svc, _ := newScreenerTestEnv()

	_, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType:  UniverseCustom,
		CustomTickers: nil,
		RiskProfile:   "MODERATE",
	})
	if !errors.Is(err, ErrEmptyUniverse) {
		t.Errorf("got %v, want ErrEmptyUniverse", err)
	}
}

func TestScreenInvalidRiskProfile(t *testing.T) {
	svc, _ := newScreenerTestEnv()

	_, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType:  UniverseCustom,
		CustomTickers: []string{"BBCA"},
		RiskProfile:   "INVALID",
	})
	if !errors.Is(err, valuation.ErrInvalidRisk) {
		t.Errorf("got %v, want ErrInvalidRisk", err)
	}
}

func TestScreenMissingCachedData(t *testing.T) {
	svc, _ := newScreenerTestEnv()

	results, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType: UniverseIndex,
		UniverseName: "IDX30",
		RiskProfile:  "MODERATE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All 3 tickers in IDX30 but none have data
	for _, r := range results {
		if r.StockData != nil {
			t.Errorf("expected nil StockData for %s", r.Ticker)
		}
	}
}

func TestScreenUnknownIndex(t *testing.T) {
	svc, _ := newScreenerTestEnv()

	_, err := svc.Screen(context.Background(), ScreenRequest{
		UniverseType: UniverseIndex,
		UniverseName: "UNKNOWN",
		RiskProfile:  "MODERATE",
	})
	if !errors.Is(err, ErrUnknownIndex) {
		t.Errorf("got %v, want ErrUnknownIndex", err)
	}
}

func TestScreenerListIndexNames(t *testing.T) {
	svc, _ := newScreenerTestEnv()
	names := svc.ListIndexNames()
	if len(names) != 1 || names[0] != "IDX30" {
		t.Errorf("got %v, want [IDX30]", names)
	}
}

func TestScreenerListSectors(t *testing.T) {
	svc, _ := newScreenerTestEnv()
	sectors := svc.ListSectors()
	if len(sectors) == 0 {
		t.Error("expected at least one sector")
	}
}
