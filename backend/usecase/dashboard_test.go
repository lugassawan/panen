package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/transaction"
)

type dashboardFixture struct {
	svc        *DashboardService
	portfolios *mockPortfolioRepo
	holdings   *mockHoldingRepo
	stocks     *mockStockRepo
	paydays    *mockPaydayRepo
	txns       *mockTransactionHistoryRepo
	sectorReg  *mockSectorRegistry
	ctx        context.Context
}

func setupDashboardTest(t *testing.T) dashboardFixture {
	t.Helper()
	portfolios := newMockPortfolioRepo()
	holdings := newMockHoldingRepo()
	stocks := newMockStockRepo()
	paydays := newMockPaydayRepo()
	txns := newMockTransactionHistoryRepo()
	sectorReg := &mockSectorRegistry{
		data: map[string]string{
			"BBCA": "Banking",
			"BMRI": "Banking",
			"TLKM": "Telecom",
			"ASII": "Automotive",
		},
	}
	svc := NewDashboardService(portfolios, holdings, stocks, paydays, txns, sectorReg)
	return dashboardFixture{
		svc: svc, portfolios: portfolios, holdings: holdings,
		stocks: stocks, paydays: paydays, txns: txns,
		sectorReg: sectorReg, ctx: context.Background(),
	}
}

func newTestPortfolio(t *testing.T, f dashboardFixture, name string, maxStocks int) *portfolio.Portfolio {
	t.Helper()
	p := portfolio.NewPortfolio(
		"broker1",
		name,
		portfolio.ModeValue,
		portfolio.RiskProfileModerate,
		10000000,
		0,
		maxStocks,
	)
	if err := f.portfolios.Create(f.ctx, p); err != nil {
		t.Fatal(err)
	}
	return p
}

func dashSeedHolding(t *testing.T, f dashboardFixture, portfolioID, ticker string, avgBuy float64, lots int) {
	t.Helper()
	h := portfolio.NewHolding(portfolioID, ticker, avgBuy, lots)
	if err := f.holdings.Create(f.ctx, h); err != nil {
		t.Fatal(err)
	}
}

func dashSeedStock(t *testing.T, f dashboardFixture, ticker string, price float64) {
	t.Helper()
	if err := f.stocks.Upsert(f.ctx, &stock.Data{Ticker: ticker, Source: "mock", Price: price}); err != nil {
		t.Fatal(err)
	}
}

func TestDashboardOverviewEmpty(t *testing.T) {
	f := setupDashboardTest(t)
	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalMarketValue != 0 {
		t.Errorf("TotalMarketValue = %f, want 0", got.TotalMarketValue)
	}
	if len(got.Portfolios) != 0 {
		t.Errorf("Portfolios = %d, want 0", len(got.Portfolios))
	}
}

func TestDashboardOverviewSinglePortfolioWeight(t *testing.T) {
	f := setupDashboardTest(t)
	p := newTestPortfolio(t, f, "Value Portfolio", 5)
	dashSeedHolding(t, f, p.ID, "BBCA", 8000, 10)
	dashSeedStock(t, f, "BBCA", 9000)

	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Portfolios) != 1 {
		t.Fatalf("Portfolios = %d, want 1", len(got.Portfolios))
	}
	if got.Portfolios[0].Weight != 100 {
		t.Errorf("Weight = %f, want 100", got.Portfolios[0].Weight)
	}
}

func TestDashboardOverviewMultiplePortfolios(t *testing.T) {
	f := setupDashboardTest(t)
	p1 := newTestPortfolio(t, f, "Value", 5)
	p2 := portfolio.NewPortfolio(
		"broker1",
		"Dividend",
		portfolio.ModeDividend,
		portfolio.RiskProfileConservative,
		5000000,
		0,
		5,
	)
	if err := f.portfolios.Create(f.ctx, p2); err != nil {
		t.Fatal(err)
	}

	dashSeedHolding(t, f, p1.ID, "BBCA", 8000, 10)
	dashSeedHolding(t, f, p2.ID, "TLKM", 3000, 20)
	dashSeedStock(t, f, "BBCA", 9000)
	dashSeedStock(t, f, "TLKM", 3500)

	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// BBCA: MV = 9000 * 10 * 100 = 9,000,000; CB = 8000 * 10 * 100 = 8,000,000
	// TLKM: MV = 3500 * 20 * 100 = 7,000,000; CB = 3000 * 20 * 100 = 6,000,000
	wantMV := 9000000.0 + 7000000.0
	wantCB := 8000000.0 + 6000000.0
	if got.TotalMarketValue != wantMV {
		t.Errorf("TotalMarketValue = %f, want %f", got.TotalMarketValue, wantMV)
	}
	if got.TotalCostBasis != wantCB {
		t.Errorf("TotalCostBasis = %f, want %f", got.TotalCostBasis, wantCB)
	}
	if got.TotalPLAmount != wantMV-wantCB {
		t.Errorf("TotalPLAmount = %f, want %f", got.TotalPLAmount, wantMV-wantCB)
	}
	if len(got.TopGainers) != 2 {
		t.Errorf("TopGainers = %d, want 2", len(got.TopGainers))
	}
	if len(got.SectorAllocation) != 2 {
		t.Errorf("SectorAllocation = %d, want 2", len(got.SectorAllocation))
	}
}

func TestDashboardOverviewMissingPrice(t *testing.T) {
	f := setupDashboardTest(t)
	p := newTestPortfolio(t, f, "Test", 5)
	dashSeedHolding(t, f, p.ID, "BBCA", 8000, 10)

	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalPLAmount != 0 {
		t.Errorf("TotalPLAmount = %f, want 0 (fallback to avgBuyPrice)", got.TotalPLAmount)
	}
}

func TestDashboardOverviewTopMoversCapped(t *testing.T) {
	f := setupDashboardTest(t)
	p := newTestPortfolio(t, f, "Test", 20)

	tickers := []string{"A1", "A2", "A3", "A4", "A5", "A6", "A7", "B1", "B2", "B3", "B4", "B5", "B6", "B7"}
	for _, ticker := range tickers {
		dashSeedHolding(t, f, p.ID, ticker, 1000, 1)
	}
	for i, ticker := range tickers[:7] {
		dashSeedStock(t, f, ticker, float64(1100+i*100))
	}
	for i, ticker := range tickers[7:] {
		dashSeedStock(t, f, ticker, float64(900-i*100))
	}

	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.TopGainers) != 5 {
		t.Errorf("TopGainers = %d, want 5", len(got.TopGainers))
	}
	if len(got.TopLosers) != 5 {
		t.Errorf("TopLosers = %d, want 5", len(got.TopLosers))
	}
}

func TestDashboardOverviewUnknownSector(t *testing.T) {
	f := setupDashboardTest(t)
	p := newTestPortfolio(t, f, "Test", 5)
	dashSeedHolding(t, f, p.ID, "UNKNOWN", 1000, 1)
	dashSeedStock(t, f, "UNKNOWN", 1100)

	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.SectorAllocation) != 1 {
		t.Fatalf("SectorAllocation = %d, want 1", len(got.SectorAllocation))
	}
	if got.SectorAllocation[0].Label != "Other" {
		t.Errorf("sector label = %q, want %q", got.SectorAllocation[0].Label, "Other")
	}
}

func TestDashboardOverviewDividendIncome(t *testing.T) {
	f := setupDashboardTest(t)
	p := newTestPortfolio(t, f, "Test", 5)

	confirmed := payday.NewPaydayEvent("2025-01", p.ID, 500000)
	confirmed.Status = payday.StatusConfirmed
	confirmed.Actual = 480000
	if err := f.paydays.Create(f.ctx, confirmed); err != nil {
		t.Fatal(err)
	}

	scheduled := payday.NewPaydayEvent("2025-02", p.ID, 500000)
	scheduled.Actual = 500000
	if err := f.paydays.Create(f.ctx, scheduled); err != nil {
		t.Fatal(err)
	}

	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalDividendIncome != 480000 {
		t.Errorf("TotalDividendIncome = %f, want 480000", got.TotalDividendIncome)
	}
}

func TestDashboardOverviewRecentTransactionsLimited(t *testing.T) {
	f := setupDashboardTest(t)
	_ = newTestPortfolio(t, f, "Test", 5)

	for i := range 15 {
		f.txns.mu.Lock()
		f.txns.records = append(f.txns.records, transaction.Record{
			ID:   "txn-" + string(rune('A'+i)),
			Type: transaction.TypeBuy,
			Date: time.Now().Add(-time.Duration(i) * 24 * time.Hour),
		})
		f.txns.mu.Unlock()
	}

	got, err := f.svc.GetOverview(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.RecentTransactions) != 10 {
		t.Errorf("RecentTransactions = %d, want 10", len(got.RecentTransactions))
	}
}
