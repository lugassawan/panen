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

	svc := NewPortfolioService(portfolioRepo, holdingRepo, buyTxnRepo, brokerageRepo, stockRepo)
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
	if !errors.Is(err, ErrInvalidMode) {
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
	if !errors.Is(err, ErrInvalidRisk) {
		t.Errorf("Create() error = %v, want ErrInvalidRisk", err)
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
