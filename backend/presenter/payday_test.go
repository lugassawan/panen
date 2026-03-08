package presenter

import (
	"context"
	"strconv"
	"testing"

	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/transaction"
	"github.com/lugassawan/panen/backend/usecase"
)

// --- mock repos for payday ---

type mockPaydayRepo struct {
	events map[string]*payday.PaydayEvent
	byKey  map[string]*payday.PaydayEvent // key: month+":"+portfolioID
}

func newMockPaydayRepo() *mockPaydayRepo {
	return &mockPaydayRepo{
		events: make(map[string]*payday.PaydayEvent),
		byKey:  make(map[string]*payday.PaydayEvent),
	}
}

func (m *mockPaydayRepo) Create(_ context.Context, event *payday.PaydayEvent) error {
	m.events[event.ID] = event
	m.byKey[event.Month+":"+event.PortfolioID] = event
	return nil
}

func (m *mockPaydayRepo) GetByMonthAndPortfolio(
	_ context.Context,
	month, portfolioID string,
) (*payday.PaydayEvent, error) {
	e, ok := m.byKey[month+":"+portfolioID]
	if !ok {
		return nil, shared.ErrNotFound
	}
	return e, nil
}

func (m *mockPaydayRepo) ListByMonth(_ context.Context, month string) ([]*payday.PaydayEvent, error) {
	var result []*payday.PaydayEvent
	for _, e := range m.events {
		if e.Month == month {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockPaydayRepo) ListByPortfolioID(_ context.Context, portfolioID string) ([]*payday.PaydayEvent, error) {
	var result []*payday.PaydayEvent
	for _, e := range m.events {
		if e.PortfolioID == portfolioID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockPaydayRepo) Update(_ context.Context, event *payday.PaydayEvent) error {
	m.events[event.ID] = event
	m.byKey[event.Month+":"+event.PortfolioID] = event
	return nil
}

type mockCashFlowRepo struct {
	flows map[string]*payday.CashFlow
}

func newMockCashFlowRepo() *mockCashFlowRepo {
	return &mockCashFlowRepo{flows: make(map[string]*payday.CashFlow)}
}

func (m *mockCashFlowRepo) Create(_ context.Context, cf *payday.CashFlow) error {
	m.flows[cf.ID] = cf
	return nil
}

func (m *mockCashFlowRepo) ListByPortfolioID(_ context.Context, portfolioID string) ([]*payday.CashFlow, error) {
	var result []*payday.CashFlow
	for _, cf := range m.flows {
		if cf.PortfolioID == portfolioID {
			result = append(result, cf)
		}
	}
	return result, nil
}

func (m *mockCashFlowRepo) Delete(_ context.Context, id string) error {
	delete(m.flows, id)
	return nil
}

func newTestPaydayHandler(day int) *PaydayHandler {
	ctx := context.Background()
	paydayRepo := newMockPaydayRepo()
	cashFlowRepo := newMockCashFlowRepo()
	portfolioRepo := newMockPortfolioRepo()
	settingsRepo := newMockSettingsRepo()

	// Pre-configure payday day if > 0.
	if day > 0 {
		_ = settingsRepo.SetSetting(ctx, "payday_day", strconv.Itoa(day))
	}

	// Seed a portfolio with monthly addition so payday service picks it up.
	p := &portfolio.Portfolio{
		ID:                 "p1",
		Name:               "Test Portfolio",
		Mode:               portfolio.ModeValue,
		BrokerageAccountID: "b1",
		MonthlyAddition:    1000000,
	}
	_ = portfolioRepo.Create(ctx, p)

	txnRepo := &mockTxnRepo{}

	svc := usecase.NewPaydayService(paydayRepo, cashFlowRepo, portfolioRepo, settingsRepo, txnRepo)
	return NewPaydayHandler(ctx, svc)
}

// mockTxnRepo is a minimal transaction.Repository for presenter tests.
type mockTxnRepo struct{}

func (m *mockTxnRepo) List(_ context.Context, _ transaction.Filter) ([]transaction.Record, error) {
	return nil, nil
}

func (m *mockTxnRepo) Summarize(_ context.Context, _ transaction.Filter) (*transaction.Summary, error) {
	return &transaction.Summary{}, nil
}

func TestPaydayHandlerGetPaydayDayDefault(t *testing.T) {
	handler := newTestPaydayHandler(0)

	day, err := handler.GetPaydayDay()
	if err != nil {
		t.Fatalf("GetPaydayDay() error = %v", err)
	}
	if day != 0 {
		t.Errorf("day = %d, want 0 (not configured)", day)
	}
}

func TestPaydayHandlerSaveAndGetPaydayDay(t *testing.T) {
	handler := newTestPaydayHandler(0)

	if err := handler.SavePaydayDay(25); err != nil {
		t.Fatalf("SavePaydayDay() error = %v", err)
	}

	day, err := handler.GetPaydayDay()
	if err != nil {
		t.Fatalf("GetPaydayDay() error = %v", err)
	}
	if day != 25 {
		t.Errorf("day = %d, want 25", day)
	}
}

func TestPaydayHandlerSavePaydayDayInvalid(t *testing.T) {
	handler := newTestPaydayHandler(0)

	err := handler.SavePaydayDay(32)
	if err == nil {
		t.Error("expected error for day > 31")
	}
}

func TestPaydayHandlerGetCurrentMonthStatusNotConfigured(t *testing.T) {
	handler := newTestPaydayHandler(0)

	resp, err := handler.GetCurrentMonthStatus()
	if err != nil {
		t.Fatalf("GetCurrentMonthStatus() error = %v", err)
	}
	if resp != nil {
		t.Errorf("expected nil response when payday not configured, got %+v", resp)
	}
}

func TestPaydayHandlerGetCurrentMonthStatusConfigured(t *testing.T) {
	handler := newTestPaydayHandler(25)

	resp, err := handler.GetCurrentMonthStatus()
	if err != nil {
		t.Fatalf("GetCurrentMonthStatus() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response when payday configured")
	}
	if resp.PaydayDay != 25 {
		t.Errorf("PaydayDay = %d, want 25", resp.PaydayDay)
	}
	if len(resp.Portfolios) != 1 {
		t.Errorf("got %d portfolios, want 1", len(resp.Portfolios))
	}
}

func TestPaydayHandlerDeferPaydayInvalidDate(t *testing.T) {
	handler := newTestPaydayHandler(25)

	err := handler.DeferPayday("p1", "not-a-date")
	if err == nil {
		t.Error("expected error for invalid date format")
	}
}

func TestPaydayHandlerSkipPayday(t *testing.T) {
	handler := newTestPaydayHandler(25)

	// Trigger current month status to create the event (SCHEDULED).
	_, _ = handler.GetCurrentMonthStatus()

	// Skip requires PENDING status, but event is SCHEDULED.
	// The transition SCHEDULED→SKIPPED is invalid, so we expect an error.
	err := handler.SkipPayday("p1")
	if err == nil {
		t.Error("expected error for invalid transition from SCHEDULED")
	}
}
