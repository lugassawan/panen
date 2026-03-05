package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

type portfolioTestFixture struct {
	svc           *PortfolioService
	portfolioRepo *mockPortfolioRepo
	holdingRepo   *mockHoldingRepo
	buyTxnRepo    *mockBuyTxnRepo
	brokerageRepo *mockBrokerageRepo
	stockRepo     *mockStockRepo
	peakRepo      *mockPeakRepo
	acct          *brokerage.Account
	port          *portfolio.Portfolio
	ctx           context.Context
}

func setupPortfolioTest(t *testing.T) portfolioTestFixture {
	t.Helper()

	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	buyTxnRepo := newMockBuyTxnRepo()
	brokerageRepo := newMockBrokerageRepo()
	stockRepo := newMockStockRepo()
	peakRepo := newMockPeakRepo()

	svc := NewPortfolioService(portfolioRepo, holdingRepo, buyTxnRepo, brokerageRepo, stockRepo, peakRepo)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BuyFeePct: 0.15, SellFeePct: 0.25,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := brokerageRepo.Create(ctx, acct); err != nil {
		t.Fatalf("setup brokerage: %v", err)
	}

	port := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: acct.ID,
		Name:               "Test Portfolio",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := portfolioRepo.Create(ctx, port); err != nil {
		t.Fatalf("setup portfolio: %v", err)
	}

	return portfolioTestFixture{
		svc:           svc,
		portfolioRepo: portfolioRepo,
		holdingRepo:   holdingRepo,
		buyTxnRepo:    buyTxnRepo,
		brokerageRepo: brokerageRepo,
		stockRepo:     stockRepo,
		peakRepo:      peakRepo,
		acct:          acct,
		port:          port,
		ctx:           ctx,
	}
}

func TestPortfolioServiceCreateHappy(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.acct.ID,
		Name:               "New Portfolio",
		Mode:               portfolio.ModeDividend,
		RiskProfile:        portfolio.RiskProfileConservative,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := f.svc.Create(f.ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestPortfolioServiceCreateEmptyName(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID: shared.NewID(), Name: "",
		Mode: portfolio.ModeValue, RiskProfile: portfolio.RiskProfileModerate,
	}
	err := f.svc.Create(f.ctx, p)
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("Create() error = %v, want ErrEmptyName", err)
	}
}

func TestPortfolioServiceCreateInvalidMode(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID: shared.NewID(), Name: "P",
		Mode: "INVALID", RiskProfile: portfolio.RiskProfileModerate,
	}
	err := f.svc.Create(f.ctx, p)
	if !errors.Is(err, portfolio.ErrInvalidMode) {
		t.Errorf("Create() error = %v, want ErrInvalidMode", err)
	}
}

func TestPortfolioServiceCreateInvalidRisk(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID: shared.NewID(), Name: "P",
		Mode: portfolio.ModeValue, RiskProfile: "BAD",
	}
	err := f.svc.Create(f.ctx, p)
	if !errors.Is(err, portfolio.ErrInvalidRisk) {
		t.Errorf("Create() error = %v, want ErrInvalidRisk", err)
	}
}

func TestPortfolioServiceListByBrokerageAccountIDHappy(t *testing.T) {
	f := setupPortfolioTest(t)

	got, err := f.svc.ListByBrokerageAccountID(f.ctx, f.acct.ID)
	if err != nil {
		t.Fatalf("ListByBrokerageAccountID() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != f.port.ID {
		t.Errorf("ID = %q, want %q", got[0].ID, f.port.ID)
	}
}

func TestPortfolioServiceListByBrokerageAccountIDEmpty(t *testing.T) {
	f := setupPortfolioTest(t)

	got, err := f.svc.ListByBrokerageAccountID(f.ctx, "nonexistent")
	if err != nil {
		t.Fatalf("ListByBrokerageAccountID() error = %v", err)
	}
	if got == nil {
		return
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestPortfolioServiceAddHoldingNew(t *testing.T) {
	f := setupPortfolioTest(t)

	holding, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}
	if holding.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want BBCA", holding.Ticker)
	}
	if holding.Lots != 10 {
		t.Errorf("Lots = %d, want 10", holding.Lots)
	}
	if holding.AvgBuyPrice != 8500 {
		t.Errorf("AvgBuyPrice = %f, want 8500", holding.AvgBuyPrice)
	}

	// Verify buy transaction was created.
	txns, err := f.buyTxnRepo.ListByHoldingID(f.ctx, holding.ID)
	if err != nil {
		t.Fatalf("ListByHoldingID() error = %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("len(txns) = %d, want 1", len(txns))
	}
	// Fee = 8500 * 10 * 100 * 0.15 / 100 = 12750
	if txns[0].Fee != 12750 {
		t.Errorf("Fee = %f, want 12750", txns[0].Fee)
	}
}

func TestPortfolioServiceAddHoldingExisting(t *testing.T) {
	f := setupPortfolioTest(t)

	// First purchase: 10 lots at 8000.
	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8000, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() first error = %v", err)
	}

	// Second purchase: 5 lots at 9000.
	holding, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 9000, 5, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() second error = %v", err)
	}

	if holding.Lots != 15 {
		t.Errorf("Lots = %d, want 15", holding.Lots)
	}
	// Weighted avg: (8000*10 + 9000*5) / 15 = 125000/15 ≈ 8333.33
	expectedAvg := (8000.0*10 + 9000.0*5) / 15
	if holding.AvgBuyPrice != expectedAvg {
		t.Errorf("AvgBuyPrice = %f, want %f", holding.AvgBuyPrice, expectedAvg)
	}

	// Two transactions should exist.
	txns, err := f.buyTxnRepo.ListByHoldingID(f.ctx, holding.ID)
	if err != nil {
		t.Fatalf("ListByHoldingID() error = %v", err)
	}
	if len(txns) != 2 {
		t.Errorf("len(txns) = %d, want 2", len(txns))
	}
}

func TestPortfolioServiceAddHoldingValidation(t *testing.T) {
	f := setupPortfolioTest(t)

	tests := []struct {
		name      string
		portfolio string
		ticker    string
		price     float64
		lots      int
		wantErr   error
	}{
		{name: "empty portfolio id", portfolio: "", price: 100, lots: 1, ticker: "BBCA", wantErr: ErrEmptyID},
		{name: "empty ticker", portfolio: f.port.ID, ticker: "", price: 100, lots: 1, wantErr: ErrEmptyTicker},
		{name: "zero price", portfolio: f.port.ID, ticker: "BBCA", price: 0, lots: 1, wantErr: ErrInvalidPrice},
		{name: "negative lots", portfolio: f.port.ID, ticker: "BBCA", price: 100, lots: -1, wantErr: ErrInvalidLots},
		{
			name:      "not found portfolio",
			portfolio: "nonexistent",
			ticker:    "BBCA",
			price:     100,
			lots:      1,
			wantErr:   shared.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.svc.AddHolding(f.ctx, tt.portfolio, tt.ticker, tt.price, tt.lots, time.Now().UTC())
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("AddHolding() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPortfolioServiceGetDetailHappy(t *testing.T) {
	f := setupPortfolioTest(t)

	// Add a holding.
	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}

	// Seed stock data for valuation.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 8500,
		EPS: 500, BVPS: 3000, PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	p, holdings, err := f.svc.GetDetail(f.ctx, f.port.ID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if p.ID != f.port.ID {
		t.Errorf("Portfolio.ID = %q, want %q", p.ID, f.port.ID)
	}
	if len(holdings) != 1 {
		t.Fatalf("len(holdings) = %d, want 1", len(holdings))
	}
	if holdings[0].Holding.Ticker != "BBCA" {
		t.Errorf("Holding.Ticker = %q, want BBCA", holdings[0].Holding.Ticker)
	}
	if holdings[0].StockData == nil {
		t.Error("StockData should not be nil")
	}
	if holdings[0].Valuation == nil {
		t.Error("Valuation should not be nil")
	}
}

func TestPortfolioServiceGetDetailNoStockData(t *testing.T) {
	f := setupPortfolioTest(t)

	// Add a holding without stock data in repo.
	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "TLKM", 3500, 5, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}

	_, holdings, err := f.svc.GetDetail(f.ctx, f.port.ID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if len(holdings) != 1 {
		t.Fatalf("len(holdings) = %d, want 1", len(holdings))
	}
	if holdings[0].StockData != nil {
		t.Error("StockData should be nil when no stock data exists")
	}
	if holdings[0].Valuation != nil {
		t.Error("Valuation should be nil when no stock data exists")
	}
}

func TestPortfolioServiceGetDetailNotFound(t *testing.T) {
	f := setupPortfolioTest(t)

	_, _, err := f.svc.GetDetail(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetDetail() error = %v, want ErrNotFound", err)
	}
}

func TestPortfolioServiceGetByIDHappy(t *testing.T) {
	f := setupPortfolioTest(t)

	got, err := f.svc.GetByID(f.ctx, f.port.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != f.port.ID {
		t.Errorf("ID = %q, want %q", got.ID, f.port.ID)
	}
}

func TestPortfolioServiceGetByIDEmptyID(t *testing.T) {
	f := setupPortfolioTest(t)

	_, err := f.svc.GetByID(f.ctx, "")
	if !errors.Is(err, ErrEmptyID) {
		t.Errorf("GetByID() error = %v, want ErrEmptyID", err)
	}
}

func TestPortfolioServiceUpdateHappy(t *testing.T) {
	f := setupPortfolioTest(t)

	f.port.Name = "Updated Name"
	f.port.RiskProfile = portfolio.RiskProfileAggressive
	f.port.Capital = 50000000
	err := f.svc.Update(f.ctx, f.port)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ := f.portfolioRepo.GetByID(f.ctx, f.port.ID)
	if got.Name != "Updated Name" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated Name")
	}
	if got.RiskProfile != portfolio.RiskProfileAggressive {
		t.Errorf("RiskProfile = %q, want AGGRESSIVE", got.RiskProfile)
	}
}

func TestPortfolioServiceUpdateEmptyID(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{ID: "", Name: "X", Mode: portfolio.ModeValue, RiskProfile: portfolio.RiskProfileModerate}
	err := f.svc.Update(f.ctx, p)
	if !errors.Is(err, ErrEmptyID) {
		t.Errorf("Update() error = %v, want ErrEmptyID", err)
	}
}

func TestPortfolioServiceUpdateEmptyName(t *testing.T) {
	f := setupPortfolioTest(t)

	f.port.Name = ""
	err := f.svc.Update(f.ctx, f.port)
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("Update() error = %v, want ErrEmptyName", err)
	}
}

func TestPortfolioServiceUpdateInvalidRisk(t *testing.T) {
	f := setupPortfolioTest(t)

	f.port.RiskProfile = "BAD"
	err := f.svc.Update(f.ctx, f.port)
	if !errors.Is(err, portfolio.ErrInvalidRisk) {
		t.Errorf("Update() error = %v, want ErrInvalidRisk", err)
	}
}

func TestPortfolioServiceUpdateModeImmutable(t *testing.T) {
	f := setupPortfolioTest(t)

	updated := *f.port
	updated.Mode = portfolio.ModeDividend
	err := f.svc.Update(f.ctx, &updated)
	if !errors.Is(err, ErrModeImmutable) {
		t.Errorf("Update() error = %v, want ErrModeImmutable", err)
	}
}

func TestPortfolioServiceUpdateNotFound(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID: "nonexistent", Name: "X",
		Mode: portfolio.ModeValue, RiskProfile: portfolio.RiskProfileModerate,
	}
	err := f.svc.Update(f.ctx, p)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestPortfolioServiceDeleteHappy(t *testing.T) {
	f := setupPortfolioTest(t)

	err := f.svc.Delete(f.ctx, f.port.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = f.portfolioRepo.GetByID(f.ctx, f.port.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("portfolio should be deleted, got error = %v", err)
	}
}

func TestPortfolioServiceDeleteEmptyID(t *testing.T) {
	f := setupPortfolioTest(t)

	err := f.svc.Delete(f.ctx, "")
	if !errors.Is(err, ErrEmptyID) {
		t.Errorf("Delete() error = %v, want ErrEmptyID", err)
	}
}

func TestPortfolioServiceDeleteHasHoldings(t *testing.T) {
	f := setupPortfolioTest(t)

	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}

	err = f.svc.Delete(f.ctx, f.port.ID)
	if !errors.Is(err, ErrHasHoldings) {
		t.Errorf("Delete() error = %v, want ErrHasHoldings", err)
	}
}

func TestPortfolioServiceCreateDuplicateMode(t *testing.T) {
	f := setupPortfolioTest(t)

	// f.port is already VALUE; try creating another VALUE under same brokerage.
	dup := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.acct.ID,
		Name:               "Duplicate",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	err := f.svc.Create(f.ctx, dup)
	if !errors.Is(err, ErrDuplicateMode) {
		t.Errorf("Create() error = %v, want ErrDuplicateMode", err)
	}
}

func TestPortfolioServiceCreateDifferentModeOK(t *testing.T) {
	f := setupPortfolioTest(t)

	// f.port is VALUE; creating DIVIDEND should succeed.
	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.acct.ID,
		Name:               "Dividend Portfolio",
		Mode:               portfolio.ModeDividend,
		RiskProfile:        portfolio.RiskProfileConservative,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := f.svc.Create(f.ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestPortfolioServiceAddHoldingDuplicateAcrossSibling(t *testing.T) {
	f := setupPortfolioTest(t)

	// Create a sibling DIVIDEND portfolio under same brokerage.
	sib := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.acct.ID,
		Name:               "Dividend Portfolio",
		Mode:               portfolio.ModeDividend,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := f.svc.Create(f.ctx, sib); err != nil {
		t.Fatalf("Create sibling: %v", err)
	}

	// Add BBCA to the first portfolio.
	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() to first portfolio: %v", err)
	}

	// Adding BBCA to the sibling should fail.
	_, err = f.svc.AddHolding(f.ctx, sib.ID, "BBCA", 8500, 5, time.Now().UTC())
	if !errors.Is(err, ErrDuplicateHolding) {
		t.Errorf("AddHolding() to sibling error = %v, want ErrDuplicateHolding", err)
	}
}

func TestPortfolioServiceGetDetailTrailingStopValueMode(t *testing.T) {
	f := setupPortfolioTest(t)

	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}

	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 9000,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	_, holdings, err := f.svc.GetDetail(f.ctx, f.port.ID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if len(holdings) != 1 {
		t.Fatalf("len(holdings) = %d, want 1", len(holdings))
	}
	ts := holdings[0].TrailingStop
	if ts == nil {
		t.Fatal("TrailingStop should not be nil for VALUE mode")
	}
	if ts.PeakPrice != 9000 {
		t.Errorf("PeakPrice = %v, want 9000", ts.PeakPrice)
	}
	// Moderate risk profile = 13.5% stop
	if ts.StopPct != 13.5 {
		t.Errorf("StopPct = %v, want 13.5", ts.StopPct)
	}
	// StopPrice = 9000 * (1 - 13.5/100) = 7785
	if ts.StopPrice != 7785 {
		t.Errorf("StopPrice = %v, want 7785", ts.StopPrice)
	}
	if ts.Triggered {
		t.Error("Triggered should be false when price is above stop")
	}
	if len(ts.FundamentalExits) != 3 {
		t.Errorf("len(FundamentalExits) = %d, want 3", len(ts.FundamentalExits))
	}
}

func TestPortfolioServiceGetDetailTrailingStopDividendMode(t *testing.T) {
	f := setupPortfolioTest(t)

	// Create a DIVIDEND portfolio.
	divPort := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.acct.ID,
		Name:               "Dividend Portfolio",
		Mode:               portfolio.ModeDividend,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := f.portfolioRepo.Create(f.ctx, divPort); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	h := portfolio.NewHolding(divPort.ID, "TLKM", 3500, 10)
	if err := f.holdingRepo.Create(f.ctx, h); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd2", Ticker: "TLKM", Price: 3600,
		EPS: 200, BVPS: 1500, ROE: 14, DER: 0.8,
		PBV: 2.4, PER: 18,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	_, holdings, err := f.svc.GetDetail(f.ctx, divPort.ID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if len(holdings) != 1 {
		t.Fatalf("len(holdings) = %d, want 1", len(holdings))
	}
	if holdings[0].TrailingStop != nil {
		t.Error("TrailingStop should be nil for DIVIDEND mode")
	}
}

func TestPortfolioServiceGetDetailTrailingStopPeakUpdate(t *testing.T) {
	f := setupPortfolioTest(t)

	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}

	// First call with price 9000.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 9000,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := f.svc.SyncPeaks(f.ctx, f.port.ID); err != nil {
		t.Fatalf("SyncPeaks() error = %v", err)
	}
	_, holdings, _ := f.svc.GetDetail(f.ctx, f.port.ID)
	if holdings[0].TrailingStop.PeakPrice != 9000 {
		t.Errorf("initial PeakPrice = %v, want 9000", holdings[0].TrailingStop.PeakPrice)
	}

	// Second call with higher price 10000.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 10000,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := f.svc.SyncPeaks(f.ctx, f.port.ID); err != nil {
		t.Fatalf("SyncPeaks() error = %v", err)
	}
	_, holdings, _ = f.svc.GetDetail(f.ctx, f.port.ID)
	if holdings[0].TrailingStop.PeakPrice != 10000 {
		t.Errorf("updated PeakPrice = %v, want 10000", holdings[0].TrailingStop.PeakPrice)
	}

	// Third call with lower price 8000 — peak should stay at 10000.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 8000,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := f.svc.SyncPeaks(f.ctx, f.port.ID); err != nil {
		t.Fatalf("SyncPeaks() error = %v", err)
	}
	_, holdings, _ = f.svc.GetDetail(f.ctx, f.port.ID)
	ts := holdings[0].TrailingStop
	if ts.PeakPrice != 10000 {
		t.Errorf("peak should not decrease: PeakPrice = %v, want 10000", ts.PeakPrice)
	}
	// StopPrice = 10000 * (1 - 13.5/100) = 8650
	if ts.StopPrice != 8650 {
		t.Errorf("StopPrice = %v, want 8650", ts.StopPrice)
	}
	// 8000 <= 8650, so triggered
	if !ts.Triggered {
		t.Error("Triggered should be true when price is below stop")
	}
}

func TestPortfolioServiceGetDetailTrailingStopSeedsFromHigh52Week(t *testing.T) {
	f := setupPortfolioTest(t)

	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}

	// Price is 8000 but High52Week is 11000. Peak should seed from High52Week.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 8000, High52Week: 11000,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := f.svc.SyncPeaks(f.ctx, f.port.ID); err != nil {
		t.Fatalf("SyncPeaks() error = %v", err)
	}
	_, holdings, err := f.svc.GetDetail(f.ctx, f.port.ID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	ts := holdings[0].TrailingStop
	if ts == nil {
		t.Fatal("TrailingStop should not be nil")
	}
	if ts.PeakPrice != 11000 {
		t.Errorf("PeakPrice = %v, want 11000 (seeded from High52Week)", ts.PeakPrice)
	}
}

func TestGetDetailDoesNotPersistPeaks(t *testing.T) {
	f := setupPortfolioTest(t)

	_, err := f.svc.AddHolding(f.ctx, f.port.ID, "BBCA", 8500, 10, time.Now().UTC())
	if err != nil {
		t.Fatalf("AddHolding() error = %v", err)
	}

	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: "sd1", Ticker: "BBCA", Price: 9000,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 2.8, PER: 17,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// Call GetDetail without SyncPeaks — should still return trailing stop
	// from computed data but NOT persist any peaks.
	_, holdings, _ := f.svc.GetDetail(f.ctx, f.port.ID)
	if holdings[0].TrailingStop == nil {
		t.Fatal("TrailingStop should not be nil even without SyncPeaks")
	}

	// Verify no peaks were persisted.
	holdingID := holdings[0].Holding.ID
	peaks, peakErr := f.peakRepo.ListByHoldingIDs(f.ctx, []string{holdingID})
	if peakErr != nil {
		t.Fatalf("ListByHoldingIDs() error = %v", peakErr)
	}
	if len(peaks) != 0 {
		t.Errorf("expected 0 persisted peaks after GetDetail without SyncPeaks, got %d", len(peaks))
	}
}
