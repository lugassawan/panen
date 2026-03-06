package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

type checklistTestFixture struct {
	svc           *ChecklistService
	checklistRepo *mockChecklistResultRepo
	portfolioRepo *mockPortfolioRepo
	holdingRepo   *mockHoldingRepo
	brokerageRepo *mockBrokerageRepo
	stockRepo     *mockStockRepo
	acct          *brokerage.Account
	port          *portfolio.Portfolio
	ctx           context.Context
}

func setupChecklistTest(t *testing.T) checklistTestFixture {
	t.Helper()

	checklistRepo := newMockChecklistResultRepo()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	brokerageRepo := newMockBrokerageRepo()
	stockRepo := newMockStockRepo()

	svc := NewChecklistService(checklistRepo, portfolioRepo, holdingRepo, brokerageRepo, stockRepo, nil)
	ctx := context.Background()

	acct := &brokerage.Account{
		ID: shared.NewID(), ProfileID: "p1", BrokerName: "Ajaib",
		BuyFeePct: 0.15, SellFeePct: 0.25, SellTaxPct: 0.1,
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
		Capital:            100000000,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := portfolioRepo.Create(ctx, port); err != nil {
		t.Fatalf("setup portfolio: %v", err)
	}

	return checklistTestFixture{
		svc:           svc,
		checklistRepo: checklistRepo,
		portfolioRepo: portfolioRepo,
		holdingRepo:   holdingRepo,
		brokerageRepo: brokerageRepo,
		stockRepo:     stockRepo,
		acct:          acct,
		port:          port,
		ctx:           ctx,
	}
}

func seedStockData(t *testing.T, f checklistTestFixture, ticker string) {
	t.Helper()
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: shared.NewID(), Ticker: ticker, Price: 8500,
		High52Week: 9000, Low52Week: 7000,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 2.8, PER: 17, DividendYield: 4, PayoutRatio: 40,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("seed stock data: %v", err)
	}
}

func seedHolding(t *testing.T, f checklistTestFixture, ticker string, avgPrice float64, lots int) *portfolio.Holding {
	t.Helper()
	h := portfolio.NewHolding(f.port.ID, ticker, avgPrice, lots)
	if err := f.holdingRepo.Create(f.ctx, h); err != nil {
		t.Fatalf("seed holding: %v", err)
	}
	return h
}

func TestChecklistServiceEvaluateHappyPath(t *testing.T) {
	f := setupChecklistTest(t)
	seedStockData(t, f, "BBCA")

	eval, err := f.svc.Evaluate(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if eval.Action != checklist.ActionBuy {
		t.Errorf("Action = %q, want BUY", eval.Action)
	}
	if eval.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want BBCA", eval.Ticker)
	}

	// BUY has 4 auto checks + 3 manual checks = 7 total.
	autoCount := len(checklist.AutoCheckDefs(checklist.ActionBuy))
	manualCount := len(checklist.ManualCheckDefs(checklist.ActionBuy))
	expectedTotal := autoCount + manualCount
	if len(eval.Checks) != expectedTotal {
		t.Errorf("len(Checks) = %d, want %d", len(eval.Checks), expectedTotal)
	}

	// Verify auto and manual types are present.
	var autoSeen, manualSeen int
	for _, cr := range eval.Checks {
		switch cr.Type {
		case checklist.CheckTypeAuto:
			autoSeen++
		case checklist.CheckTypeManual:
			manualSeen++
		}
	}
	if autoSeen != autoCount {
		t.Errorf("auto checks = %d, want %d", autoSeen, autoCount)
	}
	if manualSeen != manualCount {
		t.Errorf("manual checks = %d, want %d", manualSeen, manualCount)
	}
}

func TestChecklistServiceEvaluateAllPassWithSuggestion(t *testing.T) {
	f := setupChecklistTest(t)

	// Seed stock data with values that will pass all auto checks for BUY.
	// Moderate thresholds: MinROE=12, MaxDER=1.0, MaxPositionPct=20.
	if err := f.stockRepo.Upsert(f.ctx, &stock.Data{
		ID: shared.NewID(), Ticker: "BBCA", Price: 100,
		High52Week: 200, Low52Week: 50,
		EPS: 500, BVPS: 3000, ROE: 18, DER: 0.5,
		PBV: 0.03, PER: 0.2, DividendYield: 4, PayoutRatio: 40,
		FetchedAt: time.Now().UTC(), Source: "mock",
	}); err != nil {
		t.Fatalf("seed stock data: %v", err)
	}

	// Pre-complete all manual checks.
	for _, def := range checklist.ManualCheckDefs(checklist.ActionBuy) {
		if err := f.svc.ToggleManualCheck(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy, def.Key, true); err != nil {
			t.Fatalf("ToggleManualCheck(%s) error = %v", def.Key, err)
		}
	}

	eval, err := f.svc.Evaluate(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if !eval.AllPassed {
		// Print details for debugging.
		for _, cr := range eval.Checks {
			t.Logf("check %s (%s): %s — %s", cr.Key, cr.Type, cr.Status, cr.Detail)
		}
		t.Fatal("AllPassed = false, want true")
	}
	if eval.Suggestion == nil {
		t.Error("Suggestion should not be nil when all checks pass")
	}
}

func TestChecklistServiceEvaluateSomeChecksFail(t *testing.T) {
	f := setupChecklistTest(t)
	seedStockData(t, f, "BBCA")

	// Manual checks are PENDING by default, so not all will pass.
	eval, err := f.svc.Evaluate(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if eval.AllPassed {
		t.Error("AllPassed = true, want false (manual checks are pending)")
	}
	if eval.Suggestion != nil {
		t.Error("Suggestion should be nil when not all checks pass")
	}
}

func TestChecklistServiceEvaluateStockDataNotFound(t *testing.T) {
	f := setupChecklistTest(t)

	// Do NOT seed stock data — should return ErrNoStockData.
	_, err := f.svc.Evaluate(f.ctx, f.port.ID, "UNKNOWN", checklist.ActionBuy)
	if !errors.Is(err, ErrNoStockData) {
		t.Errorf("Evaluate() error = %v, want ErrNoStockData", err)
	}
}

func TestChecklistServiceToggleManualCheckCreatesNew(t *testing.T) {
	f := setupChecklistTest(t)

	err := f.svc.ToggleManualCheck(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy, "no_negative_news", true)
	if err != nil {
		t.Fatalf("ToggleManualCheck() error = %v", err)
	}

	saved, err := f.checklistRepo.Get(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !saved.ManualChecks["no_negative_news"] {
		t.Error("ManualChecks[no_negative_news] = false, want true")
	}
}

func TestChecklistServiceToggleManualCheckUpdatesExisting(t *testing.T) {
	f := setupChecklistTest(t)

	// Create initial result.
	err := f.svc.ToggleManualCheck(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy, "no_negative_news", true)
	if err != nil {
		t.Fatalf("ToggleManualCheck() first error = %v", err)
	}

	// Update with another key.
	err = f.svc.ToggleManualCheck(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy, "thesis_still_valid", true)
	if err != nil {
		t.Fatalf("ToggleManualCheck() second error = %v", err)
	}

	saved, err := f.checklistRepo.Get(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !saved.ManualChecks["no_negative_news"] {
		t.Error("ManualChecks[no_negative_news] should still be true")
	}
	if !saved.ManualChecks["thesis_still_valid"] {
		t.Error("ManualChecks[thesis_still_valid] = false, want true")
	}
}

func TestChecklistServiceResetChecklistDeletes(t *testing.T) {
	f := setupChecklistTest(t)

	// Create a saved result.
	err := f.svc.ToggleManualCheck(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy, "no_negative_news", true)
	if err != nil {
		t.Fatalf("ToggleManualCheck() error = %v", err)
	}

	// Reset should delete it.
	err = f.svc.ResetChecklist(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("ResetChecklist() error = %v", err)
	}

	_, err = f.checklistRepo.Get(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Get() after reset error = %v, want ErrNotFound", err)
	}
}

func TestChecklistServiceResetChecklistNoOp(t *testing.T) {
	f := setupChecklistTest(t)

	// Reset when nothing is saved should be a no-op.
	err := f.svc.ResetChecklist(f.ctx, f.port.ID, "BBCA", checklist.ActionBuy)
	if err != nil {
		t.Fatalf("ResetChecklist() error = %v, want nil", err)
	}
}

func TestChecklistServiceAvailableActionsNoHoldingValue(t *testing.T) {
	f := setupChecklistTest(t)

	actions, err := f.svc.AvailableActions(f.ctx, f.port.ID, "BBCA")
	if err != nil {
		t.Fatalf("AvailableActions() error = %v", err)
	}
	if len(actions) != 1 || actions[0] != checklist.ActionBuy {
		t.Errorf("AvailableActions() = %v, want [BUY]", actions)
	}
}

func TestChecklistServiceAvailableActionsNoHoldingDividend(t *testing.T) {
	f := setupChecklistTest(t)

	// Create a dividend portfolio.
	divPort := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.acct.ID,
		Name:               "Dividend Portfolio",
		Mode:               portfolio.ModeDividend,
		RiskProfile:        portfolio.RiskProfileModerate,
		Capital:            50000000,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := f.portfolioRepo.Create(f.ctx, divPort); err != nil {
		t.Fatalf("create dividend portfolio: %v", err)
	}

	actions, err := f.svc.AvailableActions(f.ctx, divPort.ID, "BBCA")
	if err != nil {
		t.Fatalf("AvailableActions() error = %v", err)
	}
	if len(actions) != 1 || actions[0] != checklist.ActionBuy {
		t.Errorf("AvailableActions() = %v, want [BUY]", actions)
	}
}

func TestChecklistServiceAvailableActionsWithHoldingValue(t *testing.T) {
	f := setupChecklistTest(t)
	seedHolding(t, f, "BBCA", 8500, 10)

	actions, err := f.svc.AvailableActions(f.ctx, f.port.ID, "BBCA")
	if err != nil {
		t.Fatalf("AvailableActions() error = %v", err)
	}

	expected := []checklist.ActionType{
		checklist.ActionBuy,
		checklist.ActionAverageDown,
		checklist.ActionSellExit,
		checklist.ActionSellStop,
		checklist.ActionHold,
	}
	if len(actions) != len(expected) {
		t.Fatalf("len(actions) = %d, want %d", len(actions), len(expected))
	}
	for i, a := range actions {
		if a != expected[i] {
			t.Errorf("actions[%d] = %q, want %q", i, a, expected[i])
		}
	}
}

func TestChecklistServiceAvailableActionsWithHoldingDividend(t *testing.T) {
	f := setupChecklistTest(t)

	// Create a dividend portfolio with a holding.
	divPort := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.acct.ID,
		Name:               "Dividend Portfolio",
		Mode:               portfolio.ModeDividend,
		RiskProfile:        portfolio.RiskProfileModerate,
		Capital:            50000000,
		Universe:           []string{},
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := f.portfolioRepo.Create(f.ctx, divPort); err != nil {
		t.Fatalf("create dividend portfolio: %v", err)
	}

	h := portfolio.NewHolding(divPort.ID, "TLKM", 3500, 20)
	if err := f.holdingRepo.Create(f.ctx, h); err != nil {
		t.Fatalf("create holding: %v", err)
	}

	actions, err := f.svc.AvailableActions(f.ctx, divPort.ID, "TLKM")
	if err != nil {
		t.Fatalf("AvailableActions() error = %v", err)
	}

	expected := []checklist.ActionType{
		checklist.ActionAverageUp,
		checklist.ActionSellExit,
		checklist.ActionSellStop,
		checklist.ActionHold,
	}
	if len(actions) != len(expected) {
		t.Fatalf("len(actions) = %d, want %d", len(actions), len(expected))
	}
	for i, a := range actions {
		if a != expected[i] {
			t.Errorf("actions[%d] = %q, want %q", i, a, expected[i])
		}
	}
}
