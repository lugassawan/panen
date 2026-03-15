package database

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
)

func newCashFlowStatementRepo(t *testing.T) (*CashFlowStatementRepo, context.Context) {
	t.Helper()
	db := newTestDB(t)
	return NewCashFlowStatementRepo(db), context.Background()
}

func TestCashFlowStatementRepoBulkUpsertAndGet(t *testing.T) {
	repo, ctx := newCashFlowStatementRepo(t)

	stmts := []financials.CashFlowStatement{
		{
			Ticker:            "BBCA",
			Source:            "fmp",
			FiscalYear:        2024,
			Quarter:           0,
			Period:            financials.PeriodAnnual,
			FetchedAt:         time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			OperatingCashFlow: 20000000,
			FreeCashFlow:      15000000,
			DividendsPaid:     5000000,
		},
		{
			Ticker:            "BBCA",
			Source:            "fmp",
			FiscalYear:        2023,
			Quarter:           0,
			Period:            financials.PeriodAnnual,
			FetchedAt:         time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			OperatingCashFlow: 18000000,
			FreeCashFlow:      13000000,
			DividendsPaid:     4500000,
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
	if got[0].FiscalYear != 2024 {
		t.Errorf("got[0].FiscalYear = %d, want 2024", got[0].FiscalYear)
	}
	if got[0].OperatingCashFlow != 20000000 {
		t.Errorf("got[0].OperatingCashFlow = %v, want 20000000", got[0].OperatingCashFlow)
	}
	if got[1].FreeCashFlow != 13000000 {
		t.Errorf("got[1].FreeCashFlow = %v, want 13000000", got[1].FreeCashFlow)
	}
}

func TestCashFlowStatementRepoBulkUpsertUpdatesExisting(t *testing.T) {
	repo, ctx := newCashFlowStatementRepo(t)

	stmts := []financials.CashFlowStatement{
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Period: financials.PeriodAnnual,
			FetchedAt:         time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			FreeCashFlow:      15000000,
			OperatingCashFlow: 20000000,
		},
	}
	if err := repo.BulkUpsert(ctx, stmts); err != nil {
		t.Fatalf("BulkUpsert() insert error = %v", err)
	}

	stmts[0].FreeCashFlow = 16000000
	stmts[0].OperatingCashFlow = 22000000
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
	if got[0].FreeCashFlow != 16000000 {
		t.Errorf("FreeCashFlow = %v, want 16000000", got[0].FreeCashFlow)
	}
	if got[0].OperatingCashFlow != 22000000 {
		t.Errorf("OperatingCashFlow = %v, want 22000000", got[0].OperatingCashFlow)
	}
}

func TestCashFlowStatementRepoLatestFetchedAt(t *testing.T) {
	repo, ctx := newCashFlowStatementRepo(t)

	latest, err := repo.LatestFetchedAt(ctx, "BBCA", "fmp")
	if err != nil {
		t.Fatalf("LatestFetchedAt() error = %v", err)
	}
	if !latest.IsZero() {
		t.Errorf("LatestFetchedAt() = %v, want zero", latest)
	}

	stmts := []financials.CashFlowStatement{
		{
			Ticker: "BBCA", Source: "fmp", FiscalYear: 2024, Period: financials.PeriodAnnual,
			FetchedAt: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	if err := repo.BulkUpsert(ctx, stmts); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	latest, err = repo.LatestFetchedAt(ctx, "BBCA", "fmp")
	if err != nil {
		t.Fatalf("LatestFetchedAt() error = %v", err)
	}
	want := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	if !latest.Equal(want) {
		t.Errorf("LatestFetchedAt() = %v, want %v", latest, want)
	}
}

func TestCashFlowStatementRepoGetByTickerEmpty(t *testing.T) {
	repo, ctx := newCashFlowStatementRepo(t)

	got, err := repo.GetByTicker(ctx, "NONEXIST", "fmp", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}
