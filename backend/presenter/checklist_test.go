package presenter

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/usecase"
)

// --- mock checklist repo ---

type mockChecklistRepo struct {
	results map[string]*checklist.ChecklistResult // key: portfolioID+":"+ticker+":"+action
}

func newMockChecklistRepo() *mockChecklistRepo {
	return &mockChecklistRepo{results: make(map[string]*checklist.ChecklistResult)}
}

func (m *mockChecklistRepo) Upsert(_ context.Context, r *checklist.ChecklistResult) error {
	m.results[r.PortfolioID+":"+r.Ticker+":"+string(r.Action)] = r
	return nil
}

func (m *mockChecklistRepo) Get(
	_ context.Context,
	portfolioID, ticker string,
	action checklist.ActionType,
) (*checklist.ChecklistResult, error) {
	r, ok := m.results[portfolioID+":"+ticker+":"+string(action)]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return r, nil
}

func (m *mockChecklistRepo) Delete(_ context.Context, id string) error {
	for k, r := range m.results {
		if r.ID == id {
			delete(m.results, k)
			return nil
		}
	}
	return shared.ErrNotFound
}

func (m *mockChecklistRepo) DeleteByPortfolioID(_ context.Context, portfolioID string) error {
	for k, r := range m.results {
		if r.PortfolioID == portfolioID {
			delete(m.results, k)
		}
	}
	return nil
}

func newTestChecklistHandler() *ChecklistHandler {
	ctx := context.Background()
	checklistRepo := newMockChecklistRepo()
	portfolioRepo := newMockPortfolioRepo()
	holdingRepo := newMockHoldingRepo()
	brokerageRepo := newMockBrokerageRepo()
	stockRepo := newMockStockRepo()

	// Seed brokerage account, portfolio, holding, and stock data.
	acct := &brokerage.Account{
		ID:        "b1",
		ProfileID: "profile-1",
		BuyFeePct: 0.15,
	}
	_ = brokerageRepo.Create(ctx, acct)

	p := &portfolio.Portfolio{
		ID:                 "p1",
		Name:               "Test Portfolio",
		Mode:               portfolio.ModeValue,
		BrokerageAccountID: "b1",
		RiskProfile:        portfolio.RiskProfileModerate,
	}
	_ = portfolioRepo.Create(ctx, p)

	h := &portfolio.Holding{
		ID:          "h1",
		PortfolioID: "p1",
		Ticker:      "BBCA",
		Lots:        10,
		AvgBuyPrice: 8000,
	}
	_ = holdingRepo.Create(ctx, h)

	_ = stockRepo.Upsert(ctx, &stock.Data{
		ID:        "s1",
		Ticker:    "BBCA",
		Price:     9000,
		EPS:       500,
		BVPS:      4000,
		ROE:       12.5,
		DER:       0.8,
		PBV:       2.25,
		PER:       18,
		FetchedAt: time.Now().UTC(),
		Source:    "mock",
	})

	svc := usecase.NewChecklistService(checklistRepo, portfolioRepo, holdingRepo, brokerageRepo, stockRepo, nil)
	return NewChecklistHandler(ctx, svc)
}

func TestChecklistHandlerEvaluate(t *testing.T) {
	handler := newTestChecklistHandler()

	resp, err := handler.EvaluateChecklist("p1", "BBCA", "HOLD")
	if err != nil {
		t.Fatalf("EvaluateChecklist() error = %v", err)
	}
	if resp.Action != "HOLD" {
		t.Errorf("Action = %q, want %q", resp.Action, "HOLD")
	}
	if resp.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want %q", resp.Ticker, "BBCA")
	}
	if len(resp.Checks) == 0 {
		t.Error("expected at least one check result")
	}
}

func TestChecklistHandlerEvaluateInvalidAction(t *testing.T) {
	handler := newTestChecklistHandler()

	_, err := handler.EvaluateChecklist("p1", "BBCA", "INVALID")
	if err == nil {
		t.Error("expected error for invalid action type")
	}
}

func TestChecklistHandlerToggleManualCheck(t *testing.T) {
	handler := newTestChecklistHandler()

	// First evaluate to create the result.
	_, _ = handler.EvaluateChecklist("p1", "BBCA", "HOLD")

	err := handler.ToggleManualCheck("p1", "BBCA", "HOLD", "some_check", true)
	if err != nil {
		t.Fatalf("ToggleManualCheck() error = %v", err)
	}
}

func TestChecklistHandlerToggleManualCheckInvalidAction(t *testing.T) {
	handler := newTestChecklistHandler()

	err := handler.ToggleManualCheck("p1", "BBCA", "INVALID", "check", true)
	if err == nil {
		t.Error("expected error for invalid action type")
	}
}

func TestChecklistHandlerResetChecklist(t *testing.T) {
	handler := newTestChecklistHandler()

	// First evaluate to create the result.
	_, _ = handler.EvaluateChecklist("p1", "BBCA", "HOLD")

	err := handler.ResetChecklist("p1", "BBCA", "HOLD")
	if err != nil {
		t.Fatalf("ResetChecklist() error = %v", err)
	}
}

func TestChecklistHandlerResetChecklistInvalidAction(t *testing.T) {
	handler := newTestChecklistHandler()

	err := handler.ResetChecklist("p1", "BBCA", "INVALID")
	if err == nil {
		t.Error("expected error for invalid action type")
	}
}

func TestChecklistHandlerAvailableActions(t *testing.T) {
	handler := newTestChecklistHandler()

	actions, err := handler.AvailableActions("p1", "BBCA")
	if err != nil {
		t.Fatalf("AvailableActions() error = %v", err)
	}
	if len(actions) == 0 {
		t.Error("expected at least one available action")
	}
}
