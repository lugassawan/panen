package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

const testTicker = "BBCA"

func TestStockServiceLookupCacheMiss(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	svc := NewStockService(repo, provider)

	data, result, err := svc.Lookup(context.Background(), testTicker, valuation.RiskConservative)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if data.Ticker != testTicker {
		t.Errorf("Ticker = %q, want %s", data.Ticker, testTicker)
	}
	if data.Price != 8500 {
		t.Errorf("Price = %f, want 8500", data.Price)
	}
	if result.Ticker != testTicker {
		t.Errorf("Result.Ticker = %q, want %s", result.Ticker, testTicker)
	}
	if result.GrahamNumber == 0 {
		t.Error("GrahamNumber should be non-zero")
	}

	provider.mu.Lock()
	calls := provider.callCount
	provider.mu.Unlock()
	if calls != 1 {
		t.Errorf("FetchPrice called %d times, want 1", calls)
	}
}

func TestStockServiceLookupCacheHit(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	svc := NewStockService(repo, provider)

	// Seed the cache with fresh data.
	cached := &stock.Data{
		ID: "cached-id", Ticker: testTicker, Price: 8000,
		EPS: 500, BVPS: 3000, PBV: 2.6, PER: 16,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}
	if err := repo.Upsert(context.Background(), cached); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	data, _, err := svc.Lookup(context.Background(), testTicker, valuation.RiskConservative)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if data.Price != 8000 {
		t.Errorf("Price = %f, want 8000 (cached)", data.Price)
	}

	// Provider should NOT have been called.
	provider.mu.Lock()
	calls := provider.callCount
	provider.mu.Unlock()
	if calls != 0 {
		t.Errorf("FetchPrice called %d times, want 0 (cache hit)", calls)
	}
}

func TestStockServiceLookupCacheStale(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	svc := NewStockService(repo, provider)

	// Seed with stale data (older than 24h).
	stale := &stock.Data{
		ID: "stale-id", Ticker: testTicker, Price: 7500,
		EPS: 500, BVPS: 3000, PBV: 2.5, PER: 15,
		FetchedAt: time.Now().UTC().Add(-25 * time.Hour), Source: "mock",
	}
	if err := repo.Upsert(context.Background(), stale); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	data, _, err := svc.Lookup(context.Background(), testTicker, valuation.RiskConservative)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if data.Price != 8500 {
		t.Errorf("Price = %f, want 8500 (fresh fetch)", data.Price)
	}

	provider.mu.Lock()
	calls := provider.callCount
	provider.mu.Unlock()
	if calls != 1 {
		t.Errorf("FetchPrice called %d times, want 1 (stale refresh)", calls)
	}
}

func TestStockServiceLookupEmptyTicker(t *testing.T) {
	svc := NewStockService(newMockStockRepo(), newMockProvider())

	_, _, err := svc.Lookup(context.Background(), "  ", valuation.RiskConservative)
	if !errors.Is(err, ErrEmptyTicker) {
		t.Errorf("Lookup() error = %v, want ErrEmptyTicker", err)
	}
}

func TestStockServiceLookupNormalizeTicker(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo, newMockProvider())

	data, _, err := svc.Lookup(context.Background(), " bbca ", valuation.RiskConservative)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if data.Ticker != testTicker {
		t.Errorf("Ticker = %q, want %s (normalized)", data.Ticker, testTicker)
	}
}

func TestStockServiceGetCachedNoData(t *testing.T) {
	svc := NewStockService(newMockStockRepo(), newMockProvider())

	_, _, err := svc.GetCached(context.Background(), testTicker, valuation.RiskConservative)
	if !errors.Is(err, ErrNoStockData) {
		t.Errorf("GetCached() error = %v, want ErrNoStockData", err)
	}
}

func TestStockServiceGetCachedHappy(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo, newMockProvider())

	cached := &stock.Data{
		ID: "id", Ticker: "BBRI", Price: 4500,
		EPS: 300, BVPS: 2000, PBV: 2.2, PER: 15,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}
	if err := repo.Upsert(context.Background(), cached); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	data, result, err := svc.GetCached(context.Background(), "BBRI", valuation.RiskModerate)
	if err != nil {
		t.Fatalf("GetCached() error = %v", err)
	}
	if data.Ticker != "BBRI" {
		t.Errorf("Ticker = %q, want BBRI", data.Ticker)
	}
	if result == nil {
		t.Error("Result should not be nil")
	}
}
