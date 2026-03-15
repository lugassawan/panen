package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
)

// fsIncomeRepo is a test double for IncomeStatementRepo.
type fsIncomeRepo struct {
	stmts     []financials.IncomeStatement
	fetchedAt time.Time
	upserted  bool
}

func (m *fsIncomeRepo) BulkUpsert(_ context.Context, stmts []financials.IncomeStatement) error {
	m.stmts = stmts
	m.upserted = true
	return nil
}

func (m *fsIncomeRepo) GetByTicker(
	_ context.Context,
	_, _ string,
	_ financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	return m.stmts, nil
}

func (m *fsIncomeRepo) LatestFetchedAt(_ context.Context, _, _ string) (time.Time, error) {
	return m.fetchedAt, nil
}

// fsBalanceRepo is a test double for BalanceSheetRepo.
type fsBalanceRepo struct {
	sheets    []financials.BalanceSheet
	fetchedAt time.Time
	upserted  bool
}

func (m *fsBalanceRepo) BulkUpsert(_ context.Context, sheets []financials.BalanceSheet) error {
	m.sheets = sheets
	m.upserted = true
	return nil
}

func (m *fsBalanceRepo) GetByTicker(
	_ context.Context,
	_, _ string,
	_ financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	return m.sheets, nil
}

func (m *fsBalanceRepo) LatestFetchedAt(_ context.Context, _, _ string) (time.Time, error) {
	return m.fetchedAt, nil
}

// fsCashFlowRepo is a test double for CashFlowStatementRepo.
type fsCashFlowRepo struct {
	stmts     []financials.CashFlowStatement
	fetchedAt time.Time
	upserted  bool
}

func (m *fsCashFlowRepo) BulkUpsert(_ context.Context, stmts []financials.CashFlowStatement) error {
	m.stmts = stmts
	m.upserted = true
	return nil
}

func (m *fsCashFlowRepo) GetByTicker(
	_ context.Context,
	_, _ string,
	_ financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	return m.stmts, nil
}

func (m *fsCashFlowRepo) LatestFetchedAt(_ context.Context, _, _ string) (time.Time, error) {
	return m.fetchedAt, nil
}

// mockFSProvider is a test double for FinancialStatementProvider.
type mockFSProvider struct {
	source      string
	incomeStmts []financials.IncomeStatement
	balSheets   []financials.BalanceSheet
	cashFlows   []financials.CashFlowStatement
	err         error
	called      bool
}

func (m *mockFSProvider) Source() string { return m.source }

func (m *mockFSProvider) FetchIncomeStatements(
	_ context.Context, _ string, _ financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	m.called = true
	return m.incomeStmts, m.err
}

func (m *mockFSProvider) FetchBalanceSheets(
	_ context.Context, _ string, _ financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	m.called = true
	return m.balSheets, m.err
}

func (m *mockFSProvider) FetchCashFlowStatements(
	_ context.Context, _ string, _ financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	m.called = true
	return m.cashFlows, m.err
}

func TestGetIncomeStatementsFetchesWhenStale(t *testing.T) {
	repo := &fsIncomeRepo{}
	provider := &mockFSProvider{
		source:      "fmp",
		incomeStmts: []financials.IncomeStatement{{Ticker: "BBCA", Revenue: 50000000}},
	}

	svc := NewFinancialStatementsService(repo, &fsBalanceRepo{}, &fsCashFlowRepo{}, provider)

	stmts, err := svc.GetIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetIncomeStatements() error = %v", err)
	}
	if !provider.called {
		t.Error("expected provider to be called for stale data")
	}
	if !repo.upserted {
		t.Error("expected BulkUpsert to be called")
	}
	if len(stmts) != 1 || stmts[0].Revenue != 50000000 {
		t.Errorf("unexpected stmts: %v", stmts)
	}
}

func TestGetIncomeStatementsUsesCacheWhenFresh(t *testing.T) {
	cached := []financials.IncomeStatement{{Ticker: "BBCA", Revenue: 45000000}}
	repo := &fsIncomeRepo{
		stmts:     cached,
		fetchedAt: time.Now().UTC().Add(-1 * time.Hour), // 1h ago = fresh
	}
	provider := &mockFSProvider{source: "fmp"}

	svc := NewFinancialStatementsService(repo, &fsBalanceRepo{}, &fsCashFlowRepo{}, provider)

	stmts, err := svc.GetIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetIncomeStatements() error = %v", err)
	}
	if provider.called {
		t.Error("expected provider NOT to be called for fresh data")
	}
	if len(stmts) != 1 || stmts[0].Revenue != 45000000 {
		t.Errorf("unexpected stmts: %v", stmts)
	}
}

func TestGetIncomeStatementsProviderError(t *testing.T) {
	repo := &fsIncomeRepo{}
	provider := &mockFSProvider{
		source: "fmp",
		err:    errors.New("api down"),
	}

	svc := NewFinancialStatementsService(repo, &fsBalanceRepo{}, &fsCashFlowRepo{}, provider)

	_, err := svc.GetIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetBalanceSheetsFetchesWhenStale(t *testing.T) {
	repo := &fsBalanceRepo{}
	provider := &mockFSProvider{
		source:    "fmp",
		balSheets: []financials.BalanceSheet{{Ticker: "BBCA", TotalAssets: 1000000000}},
	}

	svc := NewFinancialStatementsService(&fsIncomeRepo{}, repo, &fsCashFlowRepo{}, provider)

	sheets, err := svc.GetBalanceSheets(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetBalanceSheets() error = %v", err)
	}
	if !provider.called {
		t.Error("expected provider to be called")
	}
	if len(sheets) != 1 || sheets[0].TotalAssets != 1000000000 {
		t.Errorf("unexpected sheets: %v", sheets)
	}
}

func TestGetCashFlowStatementsFetchesWhenStale(t *testing.T) {
	repo := &fsCashFlowRepo{}
	provider := &mockFSProvider{
		source:    "fmp",
		cashFlows: []financials.CashFlowStatement{{Ticker: "BBCA", FreeCashFlow: 15000000}},
	}

	svc := NewFinancialStatementsService(&fsIncomeRepo{}, &fsBalanceRepo{}, repo, provider)

	stmts, err := svc.GetCashFlowStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("GetCashFlowStatements() error = %v", err)
	}
	if !provider.called {
		t.Error("expected provider to be called")
	}
	if len(stmts) != 1 || stmts[0].FreeCashFlow != 15000000 {
		t.Errorf("unexpected stmts: %v", stmts)
	}
}

func TestIsFinancialsFresh(t *testing.T) {
	tests := []struct {
		name   string
		latest time.Time
		want   bool
	}{
		{"zero time is stale", time.Time{}, false},
		{"1 hour ago is fresh", time.Now().UTC().Add(-1 * time.Hour), true},
		{"25 hours ago is stale", time.Now().UTC().Add(-25 * time.Hour), false},
		{"23 hours ago is fresh", time.Now().UTC().Add(-23 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFinancialsFresh(tt.latest); got != tt.want {
				t.Errorf("isFinancialsFresh(%v) = %v, want %v", tt.latest, got, tt.want)
			}
		})
	}
}
