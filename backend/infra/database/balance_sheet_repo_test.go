package database

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
)

func newBalanceSheetRepo(t *testing.T) (*BalanceSheetRepo, context.Context) {
	t.Helper()
	db := newTestDB(t)
	return NewBalanceSheetRepo(db), context.Background()
}

func TestBalanceSheetRepoBulkUpsertAndGet(t *testing.T) {
	repo, ctx := newBalanceSheetRepo(t)

	sheets := []financials.BalanceSheet{
		{
			Ticker:      "BBCA",
			Source:      "fmp",
			FiscalYear:  2024,
			Quarter:     0,
			Period:      financials.PeriodAnnual,
			FetchedAt:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			TotalAssets: 1000000000,
			TotalEquity: 300000000,
			TotalDebt:   400000000,
		},
		{
			Ticker:      "BBCA",
			Source:      "fmp",
			FiscalYear:  2023,
			Quarter:     0,
			Period:      financials.PeriodAnnual,
			FetchedAt:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			TotalAssets: 900000000,
			TotalEquity: 280000000,
			TotalDebt:   350000000,
		},
	}

	if err := repo.BulkUpsert(ctx, sheets); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "fmp", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].FiscalYear != 2024 {
		t.Errorf("got[0].FiscalYear = %d, want 2024", got[0].FiscalYear)
	}
	if got[0].TotalAssets != 1000000000 {
		t.Errorf("got[0].TotalAssets = %v, want 1000000000", got[0].TotalAssets)
	}
}

func TestBalanceSheetRepoBulkUpsertUpdatesExisting(t *testing.T) {
	repo, ctx := newBalanceSheetRepo(t)

	sheets := []financials.BalanceSheet{
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Period: financials.PeriodAnnual,
			FetchedAt:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			TotalAssets: 1000000000, TotalEquity: 300000000,
		},
	}
	if err := repo.BulkUpsert(ctx, sheets); err != nil {
		t.Fatalf("BulkUpsert() insert error = %v", err)
	}

	// Update both assets and equity to verify multi-field update
	sheets[0].TotalAssets = 1100000000
	sheets[0].TotalEquity = 350000000
	sheets[0].ID = ""
	if err := repo.BulkUpsert(ctx, sheets); err != nil {
		t.Fatalf("BulkUpsert() update error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "fmp", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].TotalAssets != 1100000000 {
		t.Errorf("TotalAssets = %v, want 1100000000", got[0].TotalAssets)
	}
	if got[0].TotalEquity != 350000000 {
		t.Errorf("TotalEquity = %v, want 350000000", got[0].TotalEquity)
	}
}

func TestBalanceSheetRepoLatestFetchedAt(t *testing.T) {
	repo, ctx := newBalanceSheetRepo(t)

	// No data yet — should return zero time.
	latest, err := repo.LatestFetchedAt(ctx, "BBCA", "fmp")
	if err != nil {
		t.Fatalf("LatestFetchedAt() error = %v", err)
	}
	if !latest.IsZero() {
		t.Errorf("LatestFetchedAt() = %v, want zero", latest)
	}

	// Insert two periods — latest should be the most recent fetched_at across both.
	sheets := []financials.BalanceSheet{
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2023, Period: financials.PeriodAnnual,
			FetchedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Period: financials.PeriodAnnual,
			FetchedAt: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	if err := repo.BulkUpsert(ctx, sheets); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	latest, err = repo.LatestFetchedAt(ctx, "BBCA", "fmp")
	if err != nil {
		t.Fatalf("LatestFetchedAt() error = %v", err)
	}
	want := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	if !latest.Equal(want) {
		t.Errorf("LatestFetchedAt() = %v, want %v", latest, want)
	}
}

func TestBalanceSheetRepoSourceIsolation(t *testing.T) {
	repo, ctx := newBalanceSheetRepo(t)

	sheets := []financials.BalanceSheet{
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Period: financials.PeriodAnnual,
			FetchedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), TotalAssets: 1000000000,
		},
		{
			Ticker: "BBCA", Source: "sectors", FiscalYear: 2024, Period: financials.PeriodAnnual,
			FetchedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), TotalAssets: 1010000000,
		},
	}
	if err := repo.BulkUpsert(ctx, sheets); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	fmp, err := repo.GetByTicker(ctx, "BBCA", "fmp", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker(fmp) error = %v", err)
	}
	if len(fmp) != 1 || fmp[0].TotalAssets != 1000000000 {
		t.Errorf("fmp data = %v, want 1 record with TotalAssets=1000000000", fmp)
	}

	sectors, err := repo.GetByTicker(ctx, "BBCA", "sectors", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker(sectors) error = %v", err)
	}
	if len(sectors) != 1 || sectors[0].TotalAssets != 1010000000 {
		t.Errorf("sectors data = %v, want 1 record with TotalAssets=1010000000", sectors)
	}
}
