package fsprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/lugassawan/panen/backend/domain/financials"
	domainProvider "github.com/lugassawan/panen/backend/domain/provider"
)

// stubProvider is a test double for FinancialStatementProvider.
type stubProvider struct {
	source        string
	incomeStmts   []financials.IncomeStatement
	balanceSheets []financials.BalanceSheet
	cashFlows     []financials.CashFlowStatement
	err           error
}

func (s *stubProvider) Source() string { return s.source }

func (s *stubProvider) FetchIncomeStatements(
	_ context.Context, _ string, _ financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	return s.incomeStmts, s.err
}

func (s *stubProvider) FetchBalanceSheets(
	_ context.Context, _ string, _ financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	return s.balanceSheets, s.err
}

func (s *stubProvider) FetchCashFlowStatements(
	_ context.Context, _ string, _ financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	return s.cashFlows, s.err
}

func TestRegistryFallback(t *testing.T) {
	primary := &stubProvider{source: "fmp", err: errors.New("fmp down")}
	secondary := &stubProvider{
		source:      "sectors",
		incomeStmts: []financials.IncomeStatement{{Ticker: "BBCA", Source: "sectors"}},
	}

	reg := NewRegistry()
	reg.Register(primary, 1)
	reg.Register(secondary, 2)

	stmts, err := reg.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchIncomeStatements() error = %v", err)
	}
	if len(stmts) != 1 || stmts[0].Source != "sectors" {
		t.Errorf("expected fallback to sectors, got %v", stmts)
	}
}

func TestRegistryPrimarySuccess(t *testing.T) {
	primary := &stubProvider{
		source:      "fmp",
		incomeStmts: []financials.IncomeStatement{{Ticker: "BBCA", Source: "fmp"}},
	}
	secondary := &stubProvider{source: "sectors", err: errors.New("should not be called")}

	reg := NewRegistry()
	reg.Register(primary, 1)
	reg.Register(secondary, 2)

	stmts, err := reg.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchIncomeStatements() error = %v", err)
	}
	if stmts[0].Source != "fmp" {
		t.Errorf("expected primary (fmp), got %q", stmts[0].Source)
	}
}

func TestRegistryAllDown(t *testing.T) {
	primary := &stubProvider{source: "fmp", err: errors.New("fmp down")}
	secondary := &stubProvider{source: "sectors", err: errors.New("sectors down")}

	reg := NewRegistry()
	reg.Register(primary, 1)
	reg.Register(secondary, 2)

	_, err := reg.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err == nil {
		t.Fatal("expected error when all providers are down")
	}
}

func TestRegistryNoProviders(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if !errors.Is(err, financials.ErrNoStatements) {
		t.Errorf("expected ErrNoStatements, got %v", err)
	}
}

func TestRegistrySource(t *testing.T) {
	reg := NewRegistry()
	if got := reg.Source(); got != "fs-registry" {
		t.Errorf("empty registry Source() = %q, want fs-registry", got)
	}

	reg.Register(&stubProvider{source: "fmp"}, 1)
	if got := reg.Source(); got != "fmp" {
		t.Errorf("Source() = %q, want fmp", got)
	}
}

func TestRegistrySetEnabled(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubProvider{source: "fmp"}, 1)
	reg.Register(&stubProvider{source: "sectors"}, 2)

	// Disable secondary
	if !reg.SetEnabled("sectors", false) {
		t.Error("SetEnabled(sectors, false) returned false")
	}

	// Cannot disable last enabled
	if reg.SetEnabled("fmp", false) {
		t.Error("SetEnabled(fmp, false) should return false when it's the last enabled")
	}

	// Non-existent provider
	if reg.SetEnabled("nonexist", false) {
		t.Error("SetEnabled(nonexist) should return false")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubProvider{source: "fmp"}, 1)
	reg.Register(&stubProvider{source: "sectors"}, 2)

	infos := reg.List()
	if len(infos) != 2 {
		t.Fatalf("len(List()) = %d, want 2", len(infos))
	}
	if infos[0].Name != "fmp" || infos[0].Priority != 1 {
		t.Errorf("infos[0] = %+v, want fmp/1", infos[0])
	}
	if infos[1].Name != "sectors" || infos[1].Priority != 2 {
		t.Errorf("infos[1] = %+v, want sectors/2", infos[1])
	}
}

func TestRegistryHealthCheckAll(t *testing.T) {
	healthy := &stubProvider{
		source:      "fmp",
		incomeStmts: []financials.IncomeStatement{{Ticker: "BBCA"}},
	}
	down := &stubProvider{source: "sectors", err: errors.New("down")}

	reg := NewRegistry()
	reg.Register(healthy, 1)
	reg.Register(down, 2)

	reg.HealthCheckAll(context.Background())

	infos := reg.List()
	if infos[0].Status != domainProvider.StatusHealthy {
		t.Errorf("fmp status = %q, want healthy", infos[0].Status)
	}
	if infos[1].Status != domainProvider.StatusDown {
		t.Errorf("sectors status = %q, want down", infos[1].Status)
	}
}

func TestRegistryFetchBalanceSheetsFallback(t *testing.T) {
	primary := &stubProvider{source: "fmp", err: errors.New("fmp down")}
	secondary := &stubProvider{
		source:        "sectors",
		balanceSheets: []financials.BalanceSheet{{Ticker: "BBCA", Source: "sectors"}},
	}

	reg := NewRegistry()
	reg.Register(primary, 1)
	reg.Register(secondary, 2)

	sheets, err := reg.FetchBalanceSheets(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchBalanceSheets() error = %v", err)
	}
	if len(sheets) != 1 || sheets[0].Source != "sectors" {
		t.Errorf("expected fallback to sectors, got %v", sheets)
	}
}

func TestRegistryFetchCashFlowStatementsFallback(t *testing.T) {
	primary := &stubProvider{source: "fmp", err: errors.New("fmp down")}
	secondary := &stubProvider{
		source:    "sectors",
		cashFlows: []financials.CashFlowStatement{{Ticker: "BBCA", Source: "sectors"}},
	}

	reg := NewRegistry()
	reg.Register(primary, 1)
	reg.Register(secondary, 2)

	stmts, err := reg.FetchCashFlowStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchCashFlowStatements() error = %v", err)
	}
	if len(stmts) != 1 || stmts[0].Source != "sectors" {
		t.Errorf("expected fallback to sectors, got %v", stmts)
	}
}
