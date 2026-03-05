package presenter

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/crashplaybook"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/usecase"
)

// --- mock repos for crash playbook ---

type mockCrashCapitalRepo struct {
	items map[string]*crashplaybook.CrashCapital
}

func newMockCrashCapitalRepo() *mockCrashCapitalRepo {
	return &mockCrashCapitalRepo{items: make(map[string]*crashplaybook.CrashCapital)}
}

func (m *mockCrashCapitalRepo) Upsert(_ context.Context, cc *crashplaybook.CrashCapital) error {
	m.items[cc.PortfolioID] = cc
	return nil
}

func (m *mockCrashCapitalRepo) GetByPortfolioID(
	_ context.Context,
	portfolioID string,
) (*crashplaybook.CrashCapital, error) {
	cc, ok := m.items[portfolioID]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return cc, nil
}

type mockTickerCollector struct{}

func (m *mockTickerCollector) CollectAll(_ context.Context) ([]string, error) { return nil, nil }

type mockEventEmitter struct{}

func (m *mockEventEmitter) Emit(_ string, _ any) {}

func newTestCrashPlaybookHandler() *CrashPlaybookHandler {
	ctx := context.Background()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	stockRepo := newMockStockRepo()
	crashCapRepo := newMockCrashCapitalRepo()
	settingsRepo := newMockSettingsRepo()
	provider := &mockDataProvider{source: "mock"}

	refreshSvc := usecase.NewRefreshService(
		stockRepo,
		provider,
		settingsRepo,
		&mockTickerCollector{},
		&mockEventEmitter{},
	)
	svc := usecase.NewCrashPlaybookService(
		stockRepo,
		provider,
		portfolioRepo,
		holdingRepo,
		crashCapRepo,
		settingsRepo,
		refreshSvc,
	)

	// Seed a portfolio for tests.
	p := &portfolio.Portfolio{
		ID:                 "p1",
		Name:               "Test Portfolio",
		Mode:               portfolio.ModeValue,
		BrokerageAccountID: "b1",
		RiskProfile:        portfolio.RiskProfileModerate,
	}
	_ = portfolioRepo.Create(ctx, p)

	return NewCrashPlaybookHandler(ctx, svc, portfolioRepo)
}

func TestCrashPlaybookHandlerListAllPortfolios(t *testing.T) {
	handler := newTestCrashPlaybookHandler()

	list, err := handler.ListAllPortfolios()
	if err != nil {
		t.Fatalf("ListAllPortfolios() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d portfolios, want 1", len(list))
	}
	if list[0].ID != "p1" {
		t.Errorf("ID = %q, want %q", list[0].ID, "p1")
	}
	if list[0].Name != "Test Portfolio" {
		t.Errorf("Name = %q, want %q", list[0].Name, "Test Portfolio")
	}
}

func TestCrashPlaybookHandlerSaveAndGetCrashCapital(t *testing.T) {
	handler := newTestCrashPlaybookHandler()

	if err := handler.SaveCrashCapital("p1", 10000000); err != nil {
		t.Fatalf("SaveCrashCapital() error = %v", err)
	}

	resp, err := handler.GetCrashCapital("p1")
	if err != nil {
		t.Fatalf("GetCrashCapital() error = %v", err)
	}
	if resp.PortfolioID != "p1" {
		t.Errorf("PortfolioID = %q, want %q", resp.PortfolioID, "p1")
	}
	if resp.Amount != 10000000 {
		t.Errorf("Amount = %v, want 10000000", resp.Amount)
	}
	if resp.Deployed != 0 {
		t.Errorf("Deployed = %v, want 0", resp.Deployed)
	}
}

func TestCrashPlaybookHandlerSaveAndGetDeploymentSettings(t *testing.T) {
	handler := newTestCrashPlaybookHandler()

	if err := handler.SaveDeploymentSettings(30, 40, 30); err != nil {
		t.Fatalf("SaveDeploymentSettings() error = %v", err)
	}

	resp, err := handler.GetDeploymentSettings()
	if err != nil {
		t.Fatalf("GetDeploymentSettings() error = %v", err)
	}
	if resp.Normal != 30 {
		t.Errorf("Normal = %v, want 30", resp.Normal)
	}
	if resp.Crash != 40 {
		t.Errorf("Crash = %v, want 40", resp.Crash)
	}
	if resp.Extreme != 30 {
		t.Errorf("Extreme = %v, want 30", resp.Extreme)
	}
}

func TestCrashPlaybookHandlerSaveDeploymentSettingsInvalidSum(t *testing.T) {
	handler := newTestCrashPlaybookHandler()

	err := handler.SaveDeploymentSettings(30, 40, 40)
	if err == nil {
		t.Error("expected error for percentages not summing to 100")
	}
}

func TestNewMarketStatusResponse(t *testing.T) {
	now := time.Now().UTC()
	status := &crashplaybook.MarketStatus{
		Condition:   crashplaybook.MarketCrash,
		IHSGPrice:   6200,
		IHSGPeak:    7800,
		DrawdownPct: 20.5,
		FetchedAt:   now,
	}
	resp := newMarketStatusResponse(status)
	if resp.Condition != "CRASH" {
		t.Errorf("Condition = %q, want %q", resp.Condition, "CRASH")
	}
	if resp.IHSGPrice != 6200 {
		t.Errorf("IHSGPrice = %v, want 6200", resp.IHSGPrice)
	}
	if resp.IHSGPeak != 7800 {
		t.Errorf("IHSGPeak = %v, want 7800", resp.IHSGPeak)
	}
	if resp.DrawdownPct != 20.5 {
		t.Errorf("DrawdownPct = %v, want 20.5", resp.DrawdownPct)
	}
}

func TestNewStockPlaybookResponse(t *testing.T) {
	activeLevel := crashplaybook.LevelCrash
	sp := crashplaybook.StockPlaybook{
		Ticker:       "BBCA",
		CurrentPrice: 8500,
		EntryPrice:   9000,
		Levels: []crashplaybook.ResponseLevel{
			{Level: crashplaybook.LevelNormalDip, TriggerPrice: 8100, DeployPct: 30},
			{Level: crashplaybook.LevelCrash, TriggerPrice: 7200, DeployPct: 40},
		},
		ActiveLevel: &activeLevel,
	}
	resp := newStockPlaybookResponse(sp)
	if resp.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want %q", resp.Ticker, "BBCA")
	}
	if resp.CurrentPrice != 8500 {
		t.Errorf("CurrentPrice = %v, want 8500", resp.CurrentPrice)
	}
	if len(resp.Levels) != 2 {
		t.Fatalf("got %d levels, want 2", len(resp.Levels))
	}
	if resp.Levels[0].Level != "NORMAL_DIP" {
		t.Errorf("Levels[0].Level = %q, want %q", resp.Levels[0].Level, "NORMAL_DIP")
	}
	if resp.ActiveLevel == nil || *resp.ActiveLevel != "CRASH" {
		t.Errorf("ActiveLevel = %v, want pointer to %q", resp.ActiveLevel, "CRASH")
	}
}

func TestNewStockPlaybookResponseNilActiveLevel(t *testing.T) {
	sp := crashplaybook.StockPlaybook{
		Ticker: "BBRI",
		Levels: []crashplaybook.ResponseLevel{},
	}
	resp := newStockPlaybookResponse(sp)
	if resp.ActiveLevel != nil {
		t.Errorf("ActiveLevel = %v, want nil", resp.ActiveLevel)
	}
}

func TestNewDiagnosticResponse(t *testing.T) {
	boolTrue := true
	boolFalse := false
	d := &crashplaybook.FallingKnifeDiagnostic{
		MarketCrashed:  true,
		CompanyBadNews: &boolFalse,
		FundamentalsOK: &boolTrue,
		BelowEntry:     true,
		Signal:         crashplaybook.SignalOpportunity,
	}
	resp := newDiagnosticResponse(d)
	if !resp.MarketCrashed {
		t.Error("expected MarketCrashed = true")
	}
	if resp.CompanyBadNews == nil || *resp.CompanyBadNews {
		t.Errorf("CompanyBadNews = %v, want pointer to false", resp.CompanyBadNews)
	}
	if resp.FundamentalsOK == nil || !*resp.FundamentalsOK {
		t.Errorf("FundamentalsOK = %v, want pointer to true", resp.FundamentalsOK)
	}
	if resp.Signal != "OPPORTUNITY" {
		t.Errorf("Signal = %q, want %q", resp.Signal, "OPPORTUNITY")
	}
}
