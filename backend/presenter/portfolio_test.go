package presenter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/trailingstop"
	"github.com/lugassawan/panen/backend/usecase"
)

// --- mock repositories ---

type mockPortfolioRepo struct {
	portfolios map[string]*portfolio.Portfolio
	byAccount  map[string][]*portfolio.Portfolio
	createErr  error
	updateErr  error
	deleteErr  error
}

func newMockPortfolioRepo() *mockPortfolioRepo {
	return &mockPortfolioRepo{
		portfolios: make(map[string]*portfolio.Portfolio),
		byAccount:  make(map[string][]*portfolio.Portfolio),
	}
}

func (m *mockPortfolioRepo) Create(_ context.Context, p *portfolio.Portfolio) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.portfolios[p.ID] = p
	m.byAccount[p.BrokerageAccountID] = append(m.byAccount[p.BrokerageAccountID], p)
	return nil
}

func (m *mockPortfolioRepo) GetByID(_ context.Context, id string) (*portfolio.Portfolio, error) {
	p, ok := m.portfolios[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return p, nil
}

func (m *mockPortfolioRepo) ListAll(_ context.Context) ([]*portfolio.Portfolio, error) {
	result := make([]*portfolio.Portfolio, 0, len(m.portfolios))
	for _, p := range m.portfolios {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockPortfolioRepo) ListByBrokerageAccountID(
	_ context.Context,
	brokerageAccountID string,
) ([]*portfolio.Portfolio, error) {
	return m.byAccount[brokerageAccountID], nil
}

func (m *mockPortfolioRepo) Update(_ context.Context, p *portfolio.Portfolio) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.portfolios[p.ID] = p
	return nil
}

func (m *mockPortfolioRepo) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.portfolios, id)
	return nil
}

type mockHoldingRepo struct {
	holdings    map[string]*portfolio.Holding
	byPortfolio map[string][]*portfolio.Holding
	byKey       map[string]*portfolio.Holding // key: portfolioID+ticker
}

func newMockHoldingRepo() *mockHoldingRepo {
	return &mockHoldingRepo{
		holdings:    make(map[string]*portfolio.Holding),
		byPortfolio: make(map[string][]*portfolio.Holding),
		byKey:       make(map[string]*portfolio.Holding),
	}
}

func (m *mockHoldingRepo) Create(_ context.Context, h *portfolio.Holding) error {
	m.holdings[h.ID] = h
	m.byPortfolio[h.PortfolioID] = append(m.byPortfolio[h.PortfolioID], h)
	m.byKey[h.PortfolioID+":"+h.Ticker] = h
	return nil
}

func (m *mockHoldingRepo) GetByID(_ context.Context, id string) (*portfolio.Holding, error) {
	h, ok := m.holdings[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return h, nil
}

func (m *mockHoldingRepo) GetByPortfolioAndTicker(
	_ context.Context,
	portfolioID, ticker string,
) (*portfolio.Holding, error) {
	h, ok := m.byKey[portfolioID+":"+ticker]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return h, nil
}

func (m *mockHoldingRepo) ListByPortfolioID(_ context.Context, portfolioID string) ([]*portfolio.Holding, error) {
	return m.byPortfolio[portfolioID], nil
}

func (m *mockHoldingRepo) Update(_ context.Context, h *portfolio.Holding) error {
	m.holdings[h.ID] = h
	m.byKey[h.PortfolioID+":"+h.Ticker] = h
	return nil
}

func (m *mockHoldingRepo) Delete(_ context.Context, id string) error {
	delete(m.holdings, id)
	return nil
}

type mockBuyTxnRepo struct{}

func (m *mockBuyTxnRepo) Create(_ context.Context, _ *portfolio.BuyTransaction) error {
	return nil
}

func (m *mockBuyTxnRepo) GetByID(_ context.Context, _ string) (*portfolio.BuyTransaction, error) {
	return nil, shared.ErrNotFound
}

func (m *mockBuyTxnRepo) ListByHoldingID(_ context.Context, _ string) ([]*portfolio.BuyTransaction, error) {
	return nil, nil
}

func (m *mockBuyTxnRepo) Delete(_ context.Context, _ string) error {
	return nil
}

type mockSellTxnRepo struct{}

func (m *mockSellTxnRepo) Create(_ context.Context, _ *portfolio.SellTransaction) error {
	return nil
}

func (m *mockSellTxnRepo) GetByID(_ context.Context, _ string) (*portfolio.SellTransaction, error) {
	return nil, shared.ErrNotFound
}

func (m *mockSellTxnRepo) ListByHoldingID(_ context.Context, _ string) ([]*portfolio.SellTransaction, error) {
	return nil, nil
}

func (m *mockSellTxnRepo) Delete(_ context.Context, _ string) error {
	return nil
}

type mockBrokerageRepo struct {
	accounts map[string]*brokerage.Account
}

func newMockBrokerageRepo() *mockBrokerageRepo {
	return &mockBrokerageRepo{accounts: make(map[string]*brokerage.Account)}
}

func (m *mockBrokerageRepo) Create(_ context.Context, a *brokerage.Account) error {
	m.accounts[a.ID] = a
	return nil
}

func (m *mockBrokerageRepo) GetByID(_ context.Context, id string) (*brokerage.Account, error) {
	a, ok := m.accounts[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return a, nil
}

func (m *mockBrokerageRepo) ListByProfileID(_ context.Context, profileID string) ([]*brokerage.Account, error) {
	var result []*brokerage.Account
	for _, a := range m.accounts {
		if a.ProfileID == profileID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockBrokerageRepo) ListNonManualByProfileID(
	_ context.Context,
	profileID string,
) ([]*brokerage.Account, error) {
	var result []*brokerage.Account
	for _, a := range m.accounts {
		if a.ProfileID == profileID && !a.IsManualFee {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockBrokerageRepo) Update(_ context.Context, a *brokerage.Account) error {
	m.accounts[a.ID] = a
	return nil
}

func (m *mockBrokerageRepo) Delete(_ context.Context, id string) error {
	delete(m.accounts, id)
	return nil
}

type mockStockRepo struct {
	data map[string]*stock.Data
}

func newMockStockRepo() *mockStockRepo {
	return &mockStockRepo{data: make(map[string]*stock.Data)}
}

func (m *mockStockRepo) Upsert(_ context.Context, d *stock.Data) error {
	m.data[d.Ticker+":"+d.Source] = d
	return nil
}

func (m *mockStockRepo) GetByTicker(_ context.Context, ticker string) (*stock.Data, error) {
	for _, d := range m.data {
		if d.Ticker == ticker {
			return d, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (m *mockStockRepo) GetByTickerAndSource(_ context.Context, ticker, source string) (*stock.Data, error) {
	d, ok := m.data[ticker+":"+source]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return d, nil
}

func (m *mockStockRepo) DeleteOlderThan(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockStockRepo) ListAllTickers(_ context.Context) ([]string, error) {
	return nil, nil
}

// mockPeakRepo is an in-memory trailingstop.PeakRepository for presenter tests.
type mockPeakRepo struct {
	items map[string]*trailingstop.HoldingPeak
}

func newMockPeakRepo() *mockPeakRepo {
	return &mockPeakRepo{items: make(map[string]*trailingstop.HoldingPeak)}
}

func (r *mockPeakRepo) Upsert(_ context.Context, peak *trailingstop.HoldingPeak) error {
	r.items[peak.HoldingID] = peak
	return nil
}

func (r *mockPeakRepo) GetByHoldingID(_ context.Context, holdingID string) (*trailingstop.HoldingPeak, error) {
	hp, ok := r.items[holdingID]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return hp, nil
}

func (r *mockPeakRepo) ListByHoldingIDs(_ context.Context, holdingIDs []string) ([]*trailingstop.HoldingPeak, error) {
	var result []*trailingstop.HoldingPeak
	for _, id := range holdingIDs {
		if hp, ok := r.items[id]; ok {
			result = append(result, hp)
		}
	}
	return result, nil
}

// --- helper to build a portfolio handler with mocks ---

func newTestPortfolioHandler() (
	*PortfolioHandler, *mockPortfolioRepo, *mockHoldingRepo, *mockBrokerageRepo, *mockStockRepo,
) {
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	buyTxnRepo := &mockBuyTxnRepo{}
	sellTxnRepo := &mockSellTxnRepo{}
	brokerageRepo := newMockBrokerageRepo()
	stockRepo := newMockStockRepo()
	peakRepo := newMockPeakRepo()
	sectorRegistry := newMockSectorRegistry()

	svc := usecase.NewPortfolioService(
		portfolioRepo,
		holdingRepo,
		buyTxnRepo,
		sellTxnRepo,
		brokerageRepo,
		stockRepo,
		peakRepo,
	)
	ctx := context.Background()
	handler := NewPortfolioHandler(ctx, svc, sectorRegistry)

	return handler, portfolioRepo, holdingRepo, brokerageRepo, stockRepo
}

// --- test helpers ---

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", field, got, want)
	}
}

// --- tests ---

func TestListPortfolios(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		result, err := handler.ListPortfolios("acct-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("len = %d, want 0", len(result))
		}
	})

	t.Run("returns all portfolios for account", func(t *testing.T) {
		handler, portfolioRepo, _, _, _ := newTestPortfolioHandler()

		now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		p1 := &portfolio.Portfolio{
			ID:                 "p1",
			BrokerageAccountID: "acct-1",
			Name:               "Value",
			Mode:               portfolio.ModeValue,
			RiskProfile:        portfolio.RiskProfileModerate,
			Capital:            10000000,
			MonthlyAddition:    1000000,
			MaxStocks:          5,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		p2 := &portfolio.Portfolio{
			ID:                 "p2",
			BrokerageAccountID: "acct-1",
			Name:               "Dividend",
			Mode:               portfolio.ModeDividend,
			RiskProfile:        portfolio.RiskProfileConservative,
			Capital:            5000000,
			MonthlyAddition:    500000,
			MaxStocks:          3,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		portfolioRepo.portfolios["p1"] = p1
		portfolioRepo.portfolios["p2"] = p2
		portfolioRepo.byAccount["acct-1"] = []*portfolio.Portfolio{p1, p2}

		result, err := handler.ListPortfolios("acct-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("len = %d, want 2", len(result))
		}
		assertEqual(t, "result[0].ID", result[0].ID, "p1")
		assertEqual(t, "result[0].Name", result[0].Name, "Value")
		assertEqual(t, "result[0].Mode", result[0].Mode, "VALUE")
		assertEqual(t, "result[1].ID", result[1].ID, "p2")
		assertEqual(t, "result[1].Mode", result[1].Mode, "DIVIDEND")
	})
}

func TestCreatePortfolio(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		resp, err := handler.CreatePortfolio("acct-1", "My Value", "VALUE", "MODERATE", 10000000, 1000000, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, "Name", resp.Name, "My Value")
		assertEqual(t, "Mode", resp.Mode, "VALUE")
		assertEqual(t, "RiskProfile", resp.RiskProfile, "MODERATE")
		assertEqual(t, "Capital", resp.Capital, 10000000.0)
		assertEqual(t, "MonthlyAddition", resp.MonthlyAddition, 1000000.0)
		assertEqual(t, "MaxStocks", resp.MaxStocks, 5)
		assertEqual(t, "BrokerageAcctID", resp.BrokerageAcctID, "acct-1")
		if resp.ID == "" {
			t.Error("ID should not be empty")
		}
	})

	t.Run("invalid mode", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.CreatePortfolio("acct-1", "Bad", "INVALID", "MODERATE", 10000000, 1000000, 5)
		if err == nil {
			t.Fatal("expected error for invalid mode")
		}
		if !errors.Is(err, portfolio.ErrInvalidMode) {
			t.Errorf("error = %v, want ErrInvalidMode", err)
		}
	})

	t.Run("invalid risk profile", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.CreatePortfolio("acct-1", "Bad", "VALUE", "INVALID", 10000000, 1000000, 5)
		if err == nil {
			t.Fatal("expected error for invalid risk profile")
		}
		if !errors.Is(err, portfolio.ErrInvalidRisk) {
			t.Errorf("error = %v, want ErrInvalidRisk", err)
		}
	})

	t.Run("duplicate mode", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.CreatePortfolio("acct-1", "First", "VALUE", "MODERATE", 10000000, 1000000, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = handler.CreatePortfolio("acct-1", "Second", "VALUE", "MODERATE", 5000000, 500000, 3)
		if err == nil {
			t.Fatal("expected error for duplicate mode")
		}
		if !errors.Is(err, usecase.ErrDuplicateMode) {
			t.Errorf("error = %v, want ErrDuplicateMode", err)
		}
	})
}

func TestGetPortfolio(t *testing.T) {
	t.Run("with holdings and no valuation", func(t *testing.T) {
		handler, portfolioRepo, holdingRepo, _, _ := newTestPortfolioHandler()

		now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		p := &portfolio.Portfolio{
			ID:                 "p1",
			BrokerageAccountID: "acct-1",
			Name:               "Value",
			Mode:               portfolio.ModeValue,
			RiskProfile:        portfolio.RiskProfileModerate,
			Capital:            10000000,
			MaxStocks:          5,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		portfolioRepo.portfolios["p1"] = p
		portfolioRepo.byAccount["acct-1"] = []*portfolio.Portfolio{p}

		h := &portfolio.Holding{
			ID:          "h1",
			PortfolioID: "p1",
			Ticker:      "BBCA",
			AvgBuyPrice: 9000,
			Lots:        10,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		holdingRepo.holdings["h1"] = h
		holdingRepo.byPortfolio["p1"] = []*portfolio.Holding{h}
		holdingRepo.byKey["p1:BBCA"] = h

		resp, err := handler.GetPortfolio("p1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, "Portfolio.ID", resp.Portfolio.ID, "p1")
		assertEqual(t, "Portfolio.Name", resp.Portfolio.Name, "Value")
		if len(resp.Holdings) != 1 {
			t.Fatalf("len(Holdings) = %d, want 1", len(resp.Holdings))
		}
		assertEqual(t, "Holdings[0].Ticker", resp.Holdings[0].Ticker, "BBCA")
		assertEqual(t, "Holdings[0].AvgBuyPrice", resp.Holdings[0].AvgBuyPrice, 9000.0)
		assertEqual(t, "Holdings[0].Lots", resp.Holdings[0].Lots, 10)
		// No stock data loaded, so valuation fields should be nil
		if resp.Holdings[0].CurrentPrice != nil {
			t.Errorf("Holdings[0].CurrentPrice should be nil")
		}
		if resp.Holdings[0].Verdict != nil {
			t.Errorf("Holdings[0].Verdict should be nil")
		}
	})

	t.Run("with holdings and stock data", func(t *testing.T) {
		handler, portfolioRepo, holdingRepo, _, stockRepo := newTestPortfolioHandler()

		now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		p := &portfolio.Portfolio{
			ID:                 "p1",
			BrokerageAccountID: "acct-1",
			Name:               "Value",
			Mode:               portfolio.ModeValue,
			RiskProfile:        portfolio.RiskProfileModerate,
			Capital:            10000000,
			MaxStocks:          5,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		portfolioRepo.portfolios["p1"] = p

		h := &portfolio.Holding{
			ID:          "h1",
			PortfolioID: "p1",
			Ticker:      "BBCA",
			AvgBuyPrice: 9000,
			Lots:        10,
		}
		holdingRepo.holdings["h1"] = h
		holdingRepo.byPortfolio["p1"] = []*portfolio.Holding{h}

		stockRepo.data["BBCA:test"] = &stock.Data{
			ID:        "s1",
			Ticker:    "BBCA",
			Price:     9500,
			EPS:       500,
			BVPS:      4000,
			PBV:       2.375,
			PER:       19,
			FetchedAt: now,
			Source:    "test",
		}

		resp, err := handler.GetPortfolio("p1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Holdings) != 1 {
			t.Fatalf("len(Holdings) = %d, want 1", len(resp.Holdings))
		}
		h0 := resp.Holdings[0]
		if h0.CurrentPrice == nil {
			t.Fatal("Holdings[0].CurrentPrice should not be nil")
		}
		assertEqual(t, "Holdings[0].CurrentPrice", *h0.CurrentPrice, 9500.0)
		if h0.GrahamNumber == nil {
			t.Fatal("Holdings[0].GrahamNumber should not be nil")
		}
		if h0.Verdict == nil {
			t.Fatal("Holdings[0].Verdict should not be nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.GetPortfolio("nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent portfolio")
		}
		if !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("error = %v, want ErrNotFound", err)
		}
	})
}

func TestAddHolding(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, portfolioRepo, _, brokerageRepo, _ := newTestPortfolioHandler()

		now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		acct := &brokerage.Account{
			ID:         "acct-1",
			BrokerName: "Ajaib",
			BuyFeePct:  0.15,
		}
		brokerageRepo.accounts["acct-1"] = acct

		p := &portfolio.Portfolio{
			ID:                 "p1",
			BrokerageAccountID: "acct-1",
			Name:               "Value",
			Mode:               portfolio.ModeValue,
			RiskProfile:        portfolio.RiskProfileModerate,
			Capital:            10000000,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		portfolioRepo.portfolios["p1"] = p
		portfolioRepo.byAccount["acct-1"] = []*portfolio.Portfolio{p}

		resp, err := handler.AddHolding("p1", "BBCA", 9000, 10, "2025-06-01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Ticker != "BBCA" {
			t.Errorf("Ticker = %q, want BBCA", resp.Ticker)
		}
		if resp.AvgBuyPrice != 9000 {
			t.Errorf("AvgBuyPrice = %v, want 9000", resp.AvgBuyPrice)
		}
		if resp.Lots != 10 {
			t.Errorf("Lots = %d, want 10", resp.Lots)
		}
		if resp.ID == "" {
			t.Error("ID should not be empty")
		}
	})

	t.Run("invalid date format", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.AddHolding("p1", "BBCA", 9000, 10, "not-a-date")
		if err == nil {
			t.Fatal("expected error for invalid date")
		}
	})

	t.Run("portfolio not found", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.AddHolding("nonexistent", "BBCA", 9000, 10, "2025-06-01")
		if err == nil {
			t.Fatal("expected error for nonexistent portfolio")
		}
	})
}

func TestUpdatePortfolio(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, portfolioRepo, _, _, _ := newTestPortfolioHandler()

		now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		p := &portfolio.Portfolio{
			ID:                 "p1",
			BrokerageAccountID: "acct-1",
			Name:               "Old Name",
			Mode:               portfolio.ModeValue,
			RiskProfile:        portfolio.RiskProfileModerate,
			Capital:            10000000,
			MonthlyAddition:    1000000,
			MaxStocks:          5,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		portfolioRepo.portfolios["p1"] = p

		resp, err := handler.UpdatePortfolio("p1", "New Name", "AGGRESSIVE", 20000000, 2000000, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, "Name", resp.Name, "New Name")
		assertEqual(t, "RiskProfile", resp.RiskProfile, "AGGRESSIVE")
		assertEqual(t, "Capital", resp.Capital, 20000000.0)
		assertEqual(t, "MonthlyAddition", resp.MonthlyAddition, 2000000.0)
		assertEqual(t, "MaxStocks", resp.MaxStocks, 10)
		// Mode should remain unchanged
		assertEqual(t, "Mode", resp.Mode, "VALUE")
	})

	t.Run("invalid risk profile", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.UpdatePortfolio("p1", "Name", "INVALID", 10000000, 1000000, 5)
		if err == nil {
			t.Fatal("expected error for invalid risk profile")
		}
		if !errors.Is(err, portfolio.ErrInvalidRisk) {
			t.Errorf("error = %v, want ErrInvalidRisk", err)
		}
	})

	t.Run("portfolio not found", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		_, err := handler.UpdatePortfolio("nonexistent", "Name", "MODERATE", 10000000, 1000000, 5)
		if err == nil {
			t.Fatal("expected error for nonexistent portfolio")
		}
		if !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("error = %v, want ErrNotFound", err)
		}
	})
}

func TestDeletePortfolio(t *testing.T) {
	t.Run("success with no holdings", func(t *testing.T) {
		handler, portfolioRepo, _, _, _ := newTestPortfolioHandler()

		now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		p := &portfolio.Portfolio{
			ID:                 "p1",
			BrokerageAccountID: "acct-1",
			Name:               "Value",
			Mode:               portfolio.ModeValue,
			RiskProfile:        portfolio.RiskProfileModerate,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		portfolioRepo.portfolios["p1"] = p

		err := handler.DeletePortfolio("p1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("fails with holdings", func(t *testing.T) {
		handler, portfolioRepo, holdingRepo, _, _ := newTestPortfolioHandler()

		now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		p := &portfolio.Portfolio{
			ID:                 "p1",
			BrokerageAccountID: "acct-1",
			Name:               "Value",
			Mode:               portfolio.ModeValue,
			RiskProfile:        portfolio.RiskProfileModerate,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		portfolioRepo.portfolios["p1"] = p

		h := &portfolio.Holding{
			ID:          "h1",
			PortfolioID: "p1",
			Ticker:      "BBCA",
			AvgBuyPrice: 9000,
			Lots:        10,
		}
		holdingRepo.holdings["h1"] = h
		holdingRepo.byPortfolio["p1"] = []*portfolio.Holding{h}

		err := handler.DeletePortfolio("p1")
		if err == nil {
			t.Fatal("expected error when portfolio has holdings")
		}
		if !errors.Is(err, usecase.ErrHasHoldings) {
			t.Errorf("error = %v, want ErrHasHoldings", err)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		err := handler.DeletePortfolio("")
		if err == nil {
			t.Fatal("expected error for empty id")
		}
		if !errors.Is(err, usecase.ErrEmptyID) {
			t.Errorf("error = %v, want ErrEmptyID", err)
		}
	})
}

func TestGetHoldingSectors(t *testing.T) {
	t.Run("returns sector for known tickers", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		result := handler.GetHoldingSectors([]string{"BBCA", "TLKM"})
		assertEqual(t, "BBCA sector", result["BBCA"], "Financials")
		assertEqual(t, "TLKM sector", result["TLKM"], "Communication Services")
	})

	t.Run("returns empty string for unknown ticker", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		result := handler.GetHoldingSectors([]string{"XXXX"})
		assertEqual(t, "XXXX sector", result["XXXX"], "")
	})

	t.Run("returns empty map for empty input", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		result := handler.GetHoldingSectors([]string{})
		if len(result) != 0 {
			t.Errorf("len = %d, want 0", len(result))
		}
	})

	t.Run("returns nil map for nil input", func(t *testing.T) {
		handler, _, _, _, _ := newTestPortfolioHandler()

		result := handler.GetHoldingSectors(nil)
		if len(result) != 0 {
			t.Errorf("len = %d, want 0", len(result))
		}
	})
}
