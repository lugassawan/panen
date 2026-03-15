package database

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
)

func newIncomeStatementRepo(t *testing.T) (*IncomeStatementRepo, context.Context) {
	t.Helper()
	db := newTestDB(t)
	return NewIncomeStatementRepo(db), context.Background()
}

func TestIncomeStatementRepoBulkUpsertAndGet(t *testing.T) {
	repo, ctx := newIncomeStatementRepo(t)

	stmts := []financials.IncomeStatement{
		{
			Ticker:     "BBCA",
			Source:     "fmp",
			FiscalYear: 2024,
			Quarter:    0,
			Period:     financials.PeriodAnnual,
			FetchedAt:  time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			Revenue:    50000000,
			NetIncome:  15000000,
			EPS:        500,
		},
		{
			Ticker:     "BBCA",
			Source:     "fmp",
			FiscalYear: 2023,
			Quarter:    0,
			Period:     financials.PeriodAnnual,
			FetchedAt:  time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			Revenue:    45000000,
			NetIncome:  13000000,
			EPS:        450,
		},
	}

	if err := repo.BulkUpsert(ctx, stmts); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "fmp", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	// Ordered by fiscal_year DESC
	if got[0].FiscalYear != 2024 {
		t.Errorf("got[0].FiscalYear = %d, want 2024", got[0].FiscalYear)
	}
	if got[0].Revenue != 50000000 {
		t.Errorf("got[0].Revenue = %v, want 50000000", got[0].Revenue)
	}
	if got[0].NetIncome != 15000000 {
		t.Errorf("got[0].NetIncome = %v, want 15000000", got[0].NetIncome)
	}
	if got[0].EPS != 500 {
		t.Errorf("got[0].EPS = %v, want 500", got[0].EPS)
	}
	if got[0].Period != financials.PeriodAnnual {
		t.Errorf("got[0].Period = %q, want annual", got[0].Period)
	}
	if got[1].EPS != 450 {
		t.Errorf("got[1].EPS = %v, want 450", got[1].EPS)
	}
}

func TestIncomeStatementRepoBulkUpsertUpdatesExisting(t *testing.T) {
	repo, ctx := newIncomeStatementRepo(t)

	stmts := []financials.IncomeStatement{
		{
			Ticker:     "BBCA",
			Source:     "fmp",
			FiscalYear: 2024,
			Quarter:    0,
			Period:     financials.PeriodAnnual,
			FetchedAt:  time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			Revenue:    50000000,
		},
	}
	if err := repo.BulkUpsert(ctx, stmts); err != nil {
		t.Fatalf("BulkUpsert() insert error = %v", err)
	}

	stmts[0].Revenue = 55000000
	stmts[0].ID = ""
	if err := repo.BulkUpsert(ctx, stmts); err != nil {
		t.Fatalf("BulkUpsert() update error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "fmp", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].Revenue != 55000000 {
		t.Errorf("Revenue = %v, want 55000000", got[0].Revenue)
	}
}

func TestIncomeStatementRepoLatestFetchedAt(t *testing.T) {
	repo, ctx := newIncomeStatementRepo(t)

	// No data yet — should return zero time.
	latest, err := repo.LatestFetchedAt(ctx, "BBCA", "fmp")
	if err != nil {
		t.Fatalf("LatestFetchedAt() error = %v", err)
	}
	if !latest.IsZero() {
		t.Errorf("LatestFetchedAt() = %v, want zero", latest)
	}

	stmts := []financials.IncomeStatement{
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2023, Period: financials.PeriodAnnual,
			FetchedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Period: financials.PeriodAnnual,
			FetchedAt: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	if err := repo.BulkUpsert(ctx, stmts); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	latest, err = repo.LatestFetchedAt(ctx, "BBCA", "fmp")
	if err != nil {
		t.Fatalf("LatestFetchedAt() error = %v", err)
	}
	want := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	if !latest.Equal(want) {
		t.Errorf("LatestFetchedAt() = %v, want %v", latest, want)
	}
}

func TestIncomeStatementRepoPeriodIsolation(t *testing.T) {
	repo, ctx := newIncomeStatementRepo(t)

	stmts := []financials.IncomeStatement{
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Quarter: 0,
			Period: financials.PeriodAnnual, FetchedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			Revenue: 50000000,
		},
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Quarter: 1,
			Period: financials.PeriodQuarterly, FetchedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			Revenue: 12000000,
		},
	}
	if err := repo.BulkUpsert(ctx, stmts); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	annual, err := repo.GetByTicker(ctx, "BBCA", "fmp", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker(annual) error = %v", err)
	}
	if len(annual) != 1 || annual[0].Revenue != 50000000 {
		t.Errorf("annual = %v, want 1 record with Revenue=50000000", annual)
	}

	quarterly, err := repo.GetByTicker(ctx, "BBCA", "fmp", financials.PeriodQuarterly)
	if err != nil {
		t.Fatalf("GetByTicker(quarterly) error = %v", err)
	}
	if len(quarterly) != 1 || quarterly[0].Revenue != 12000000 {
		t.Errorf("quarterly = %v, want 1 record with Revenue=12000000", quarterly)
	}
}
