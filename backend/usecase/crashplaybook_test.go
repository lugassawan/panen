package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/crashplaybook"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

// mockCrashCapitalRepo is an in-memory crashplaybook.CrashCapitalRepository for testing.
type mockCrashCapitalRepo struct {
	mu    sync.Mutex
	items map[string]*crashplaybook.CrashCapital // keyed by portfolioID
}

func newMockCrashCapitalRepo() *mockCrashCapitalRepo {
	return &mockCrashCapitalRepo{items: make(map[string]*crashplaybook.CrashCapital)}
}

func (r *mockCrashCapitalRepo) Upsert(_ context.Context, cc *crashplaybook.CrashCapital) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[cc.PortfolioID] = cc
	return nil
}

func (r *mockCrashCapitalRepo) GetByPortfolioID(
	_ context.Context,
	portfolioID string,
) (*crashplaybook.CrashCapital, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cc, ok := r.items[portfolioID]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return cc, nil
}

func newTestCrashPlaybookService() (
	*CrashPlaybookService, *mockStockRepo, *mockPortfolioRepo,
	*mockHoldingRepo, *mockCrashCapitalRepo,
) {
	stockRepo := newMockStockRepo()
	provider := newMockProvider()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	ccRepo := newMockCrashCapitalRepo()
	settingsRepo := newMockSettingsRepo()

	// Seed default deployment settings.
	settingsRepo.kv["crash_deploy_pct_normal"] = "30"
	settingsRepo.kv["crash_deploy_pct_crash"] = "40"
	settingsRepo.kv["crash_deploy_pct_extreme"] = "30"

	svc := NewCrashPlaybookService(stockRepo, provider, portfolioRepo, holdingRepo, ccRepo, settingsRepo, nil)
	return svc, stockRepo, portfolioRepo, holdingRepo, ccRepo
}

func TestGetMarketStatus(t *testing.T) {
	svc, stockRepo, _, _, _ := newTestCrashPlaybookService()
	ctx := context.Background()

	// Seed IHSG data — price within 5% of peak so condition is NORMAL.
	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID:         "ihsg-1",
		Ticker:     "^JKSE",
		Price:      7200,
		High52Week: 7500,
		Low52Week:  6000,
		FetchedAt:  time.Now().UTC(),
		Source:     "mock",
	})

	status, err := svc.GetMarketStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Condition != crashplaybook.MarketNormal {
		t.Errorf("expected NORMAL, got %v", status.Condition)
	}
	if status.IHSGPrice != 7200 {
		t.Errorf("expected price 7200, got %v", status.IHSGPrice)
	}
}

func TestGetMarketStatusCrashCondition(t *testing.T) {
	svc, stockRepo, _, _, _ := newTestCrashPlaybookService()
	ctx := context.Background()

	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID:         "ihsg-1",
		Ticker:     "^JKSE",
		Price:      5500,
		High52Week: 7500,
		Low52Week:  5000,
		FetchedAt:  time.Now().UTC(),
		Source:     "mock",
	})

	status, err := svc.GetMarketStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Condition != crashplaybook.MarketCrash {
		t.Errorf("expected CRASH, got %v", status.Condition)
	}
}

func TestGetPortfolioPlaybook(t *testing.T) {
	svc, stockRepo, portfolioRepo, holdingRepo, _ := newTestCrashPlaybookService()
	ctx := context.Background()

	// Seed IHSG — price within 5% of peak so condition is NORMAL.
	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID: "ihsg-1", Ticker: "^JKSE", Price: 7200, High52Week: 7500, Low52Week: 6000,
		FetchedAt: time.Now().UTC(), Source: "mock",
	})

	// Seed portfolio + holding + stock data.
	_ = portfolioRepo.Create(ctx, &portfolio.Portfolio{
		ID: "p1", BrokerageAccountID: "b1", Name: "Test",
		Mode: "VALUE", RiskProfile: "CONSERVATIVE",
	})
	_ = holdingRepo.Create(ctx, &portfolio.Holding{
		ID: "h1", PortfolioID: "p1", Ticker: "BBCA", AvgBuyPrice: 8000, Lots: 10,
	})
	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID: "bbca-1", Ticker: "BBCA", Price: 8500, High52Week: 9000, Low52Week: 7000,
		EPS: 500, BVPS: 3000, FetchedAt: time.Now().UTC(), Source: "mock",
	})

	pb, err := svc.GetPortfolioPlaybook(ctx, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pb.Market.Condition != crashplaybook.MarketNormal {
		t.Errorf("expected NORMAL market, got %v", pb.Market.Condition)
	}
	if len(pb.Stocks) != 1 {
		t.Fatalf("expected 1 stock playbook, got %d", len(pb.Stocks))
	}
	if pb.Stocks[0].Ticker != "BBCA" {
		t.Errorf("expected ticker BBCA, got %v", pb.Stocks[0].Ticker)
	}
	if len(pb.Stocks[0].Levels) != 3 {
		t.Errorf("expected 3 levels, got %d", len(pb.Stocks[0].Levels))
	}
}

func TestSaveCrashCapital(t *testing.T) {
	svc, _, _, _, ccRepo := newTestCrashPlaybookService()
	ctx := context.Background()

	if err := svc.SaveCrashCapital(ctx, "p1", 10_000_000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cc, err := ccRepo.GetByPortfolioID(ctx, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc.Amount != 10_000_000 {
		t.Errorf("expected amount 10000000, got %v", cc.Amount)
	}

	// Update existing.
	if err := svc.SaveCrashCapital(ctx, "p1", 20_000_000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cc, _ = ccRepo.GetByPortfolioID(ctx, "p1")
	if cc.Amount != 20_000_000 {
		t.Errorf("expected amount 20000000, got %v", cc.Amount)
	}
}

func TestGetDeploymentPlan(t *testing.T) {
	svc, _, _, _, _ := newTestCrashPlaybookService()
	ctx := context.Background()

	_ = svc.SaveCrashCapital(ctx, "p1", 10_000_000)

	plan, err := svc.GetDeploymentPlan(ctx, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Total != 10_000_000 {
		t.Errorf("expected total 10000000, got %v", plan.Total)
	}
	if len(plan.Levels) != 3 {
		t.Fatalf("expected 3 levels, got %d", len(plan.Levels))
	}
	if plan.Levels[0].Amount != 3_000_000 {
		t.Errorf("expected normal dip amount 3000000, got %v", plan.Levels[0].Amount)
	}
	if plan.Levels[1].Amount != 4_000_000 {
		t.Errorf("expected crash amount 4000000, got %v", plan.Levels[1].Amount)
	}
	if plan.Levels[2].Amount != 3_000_000 {
		t.Errorf("expected extreme amount 3000000, got %v", plan.Levels[2].Amount)
	}
}

func TestSaveDeploymentSettingsValidSum(t *testing.T) {
	svc, _, _, _, _ := newTestCrashPlaybookService()
	ctx := context.Background()

	if err := svc.SaveDeploymentSettings(ctx, 20, 50, 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	normal, crash, extreme, err := svc.GetDeploymentSettings(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if normal != 20 || crash != 50 || extreme != 30 {
		t.Errorf("expected 20/50/30, got %v/%v/%v", normal, crash, extreme)
	}
}

func TestSaveDeploymentSettingsInvalidSum(t *testing.T) {
	svc, _, _, _, _ := newTestCrashPlaybookService()
	ctx := context.Background()

	err := svc.SaveDeploymentSettings(ctx, 20, 50, 20)
	if !errors.Is(err, ErrInvalidDeploymentSum) {
		t.Errorf("expected ErrInvalidDeploymentSum, got %v", err)
	}
}

func TestSuggestedRefreshInterval(t *testing.T) {
	tests := []struct {
		condition crashplaybook.MarketCondition
		want      int
	}{
		{crashplaybook.MarketNormal, 720},
		{crashplaybook.MarketElevated, 360},
		{crashplaybook.MarketCorrection, 240},
		{crashplaybook.MarketCrash, 180},
		{crashplaybook.MarketRecovery, 360},
	}

	for _, tt := range tests {
		t.Run(string(tt.condition), func(t *testing.T) {
			got := SuggestedRefreshInterval(tt.condition)
			if got != tt.want {
				t.Errorf("SuggestedRefreshInterval(%v) = %v, want %v", tt.condition, got, tt.want)
			}
		})
	}
}

func TestGetDiagnostic(t *testing.T) {
	svc, stockRepo, portfolioRepo, _, _ := newTestCrashPlaybookService()
	ctx := context.Background()

	// Seed IHSG in crash condition.
	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID: "ihsg-1", Ticker: "^JKSE", Price: 5500, High52Week: 7500, Low52Week: 5000,
		FetchedAt: time.Now().UTC(), Source: "mock",
	})

	_ = portfolioRepo.Create(ctx, &portfolio.Portfolio{
		ID: "p1", BrokerageAccountID: "b1", Name: "Test",
		Mode: "VALUE", RiskProfile: "CONSERVATIVE",
	})

	// Seed stock data where price is below entry.
	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID: "bbca-1", Ticker: "BBCA", Price: 500, High52Week: 9000, Low52Week: 400,
		EPS: 500, BVPS: 3000, FetchedAt: time.Now().UTC(), Source: "mock",
	})

	boolPtr := func(v bool) *bool { return &v }

	diag, err := svc.GetDiagnostic(ctx, "BBCA", "p1", boolPtr(false), boolPtr(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !diag.MarketCrashed {
		t.Error("expected marketCrashed to be true")
	}
	if !diag.BelowEntry {
		t.Error("expected belowEntry to be true")
	}
	if diag.Signal != crashplaybook.SignalOpportunity {
		t.Errorf("expected OPPORTUNITY signal, got %v", diag.Signal)
	}
}
