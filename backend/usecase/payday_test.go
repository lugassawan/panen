package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/transaction"
)

const testExpectedAmount = 5000000

type paydayTestFixture struct {
	svc           *PaydayService
	paydayRepo    *mockPaydayRepo
	cashFlowRepo  *mockCashFlowRepo
	portfolioRepo *mockPortfolioRepo
	settingsRepo  *mockSettingsRepo
	txnRepo       *mockTransactionHistoryRepo
	ctx           context.Context
}

func setupPaydayTest(t *testing.T) paydayTestFixture {
	t.Helper()

	paydayRepo := newMockPaydayRepo()
	cashFlowRepo := newMockCashFlowRepo()
	portfolioRepo := newMockPortfolioRepo()
	settingsRepo := newMockSettingsRepo()
	txnRepo := newMockTransactionHistoryRepo()

	svc := NewPaydayService(paydayRepo, cashFlowRepo, portfolioRepo, settingsRepo, txnRepo)

	return paydayTestFixture{
		svc:           svc,
		paydayRepo:    paydayRepo,
		cashFlowRepo:  cashFlowRepo,
		portfolioRepo: portfolioRepo,
		settingsRepo:  settingsRepo,
		txnRepo:       txnRepo,
		ctx:           context.Background(),
	}
}

func addTestPortfolio(t *testing.T, f paydayTestFixture, name string, monthlyAddition float64) *portfolio.Portfolio {
	t.Helper()
	p := &portfolio.Portfolio{
		ID:              shared.NewID(),
		Name:            name,
		Mode:            portfolio.ModeValue,
		MonthlyAddition: monthlyAddition,
		Universe:        []string{},
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := f.portfolioRepo.Create(f.ctx, p); err != nil {
		t.Fatalf("setup portfolio: %v", err)
	}
	return p
}

func seedPaydayEvent(
	t *testing.T,
	f paydayTestFixture,
	month, portfolioID string,
	status payday.Status,
) {
	t.Helper()
	ev := payday.NewPaydayEvent(month, portfolioID, testExpectedAmount)
	ev.Status = status
	if err := f.paydayRepo.Create(f.ctx, ev); err != nil {
		t.Fatalf("seed payday event: %v", err)
	}
}

func TestGetPaydayDay(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(f paydayTestFixture)
		wantDay int
	}{
		{
			name:    "returns 0 when not set",
			setup:   func(_ paydayTestFixture) {},
			wantDay: 0,
		},
		{
			name: "returns configured value",
			setup: func(f paydayTestFixture) {
				if err := f.settingsRepo.SetSetting(f.ctx, "payday_day", "25"); err != nil {
					t.Fatalf("set setting: %v", err)
				}
			},
			wantDay: 25,
		},
		{
			name: "returns 0 for empty string",
			setup: func(f paydayTestFixture) {
				if err := f.settingsRepo.SetSetting(f.ctx, "payday_day", ""); err != nil {
					t.Fatalf("set setting: %v", err)
				}
			},
			wantDay: 0,
		},
		{
			name: "returns 0 for non-numeric value",
			setup: func(f paydayTestFixture) {
				if err := f.settingsRepo.SetSetting(f.ctx, "payday_day", "abc"); err != nil {
					t.Fatalf("set setting: %v", err)
				}
			},
			wantDay: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupPaydayTest(t)
			tt.setup(f)

			got, err := f.svc.GetPaydayDay(f.ctx)
			if err != nil {
				t.Fatalf("GetPaydayDay() error = %v", err)
			}
			if got != tt.wantDay {
				t.Errorf("GetPaydayDay() = %d, want %d", got, tt.wantDay)
			}
		})
	}
}

func TestSavePaydayDay(t *testing.T) {
	tests := []struct {
		name    string
		day     int
		wantErr bool
	}{
		{name: "valid day 0 (disable)", day: 0},
		{name: "valid day 1", day: 1},
		{name: "valid day 15", day: 15},
		{name: "valid day 31", day: 31},
		{name: "negative day", day: -1, wantErr: true},
		{name: "day 32", day: 32, wantErr: true},
		{name: "large negative", day: -100, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupPaydayTest(t)

			err := f.svc.SavePaydayDay(f.ctx, tt.day)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SavePaydayDay(%d) expected error, got nil", tt.day)
				} else if !errors.Is(err, ErrInvalidPaydayDay) {
					t.Errorf("SavePaydayDay(%d) error = %v, want %v", tt.day, err, ErrInvalidPaydayDay)
				}
				return
			}
			if err != nil {
				t.Fatalf("SavePaydayDay(%d) unexpected error: %v", tt.day, err)
			}

			// Verify persisted value.
			got, err := f.svc.GetPaydayDay(f.ctx)
			if err != nil {
				t.Fatalf("GetPaydayDay() error = %v", err)
			}
			if got != tt.day {
				t.Errorf("GetPaydayDay() = %d, want %d", got, tt.day)
			}
		})
	}
}

func TestGetCurrentMonthStatusNotConfigured(t *testing.T) {
	f := setupPaydayTest(t)
	// payday_day defaults to 0 (not set)

	_, err := f.svc.GetCurrentMonthStatus(f.ctx)
	if !errors.Is(err, ErrPaydayNotConfigured) {
		t.Errorf("expected ErrPaydayNotConfigured, got %v", err)
	}
}

func TestGetCurrentMonthStatusEmptyPortfolios(t *testing.T) {
	f := setupPaydayTest(t)
	if err := f.svc.SavePaydayDay(f.ctx, 1); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	// Add portfolio with 0 monthly addition.
	addTestPortfolio(t, f, "No Addition", 0)

	status, err := f.svc.GetCurrentMonthStatus(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if len(status.Portfolios) != 0 {
		t.Errorf("expected 0 portfolios, got %d", len(status.Portfolios))
	}
}

func TestGetCurrentMonthStatusLazyCreates(t *testing.T) {
	f := setupPaydayTest(t)
	// Set payday day to 1 so today >= payday day (any day of month >= 1).
	if err := f.svc.SavePaydayDay(f.ctx, 1); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	p := addTestPortfolio(t, f, "Growth", 5000000)

	status, err := f.svc.GetCurrentMonthStatus(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if len(status.Portfolios) != 1 {
		t.Fatalf("expected 1 portfolio, got %d", len(status.Portfolios))
	}
	ps := status.Portfolios[0]
	if ps.PortfolioID != p.ID {
		t.Errorf("portfolio ID = %s, want %s", ps.PortfolioID, p.ID)
	}
	if ps.Expected != 5000000 {
		t.Errorf("expected = %f, want %f", ps.Expected, 5000000.0)
	}
	// Day 1 means today >= payday day is always true, so status should be PENDING.
	if ps.Status != string(payday.StatusPending) {
		t.Errorf("status = %s, want %s", ps.Status, payday.StatusPending)
	}
}

func TestGetCurrentMonthStatusScheduled(t *testing.T) {
	f := setupPaydayTest(t)
	// Set payday day to 31 — most months, today < 31.
	if err := f.svc.SavePaydayDay(f.ctx, 31); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	addTestPortfolio(t, f, "Scheduled", 3000000)

	now := time.Now().UTC()
	// Only test if today is not the 31st.
	if now.Day() >= 31 {
		t.Skip("today is the 31st, cannot test SCHEDULED status")
	}

	status, err := f.svc.GetCurrentMonthStatus(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if len(status.Portfolios) != 1 {
		t.Fatalf("expected 1 portfolio, got %d", len(status.Portfolios))
	}
	if status.Portfolios[0].Status != string(payday.StatusScheduled) {
		t.Errorf("status = %s, want %s", status.Portfolios[0].Status, payday.StatusScheduled)
	}
}

func TestGetCurrentMonthStatusAutoTransition(t *testing.T) {
	f := setupPaydayTest(t)
	if err := f.svc.SavePaydayDay(f.ctx, 1); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	p := addTestPortfolio(t, f, "Deferred", 2000000)

	now := time.Now().UTC()
	currentMonth := now.Format("2006-01")

	// Pre-create a DEFERRED event with DeferUntil in the past.
	ev := payday.NewPaydayEvent(currentMonth, p.ID, 2000000)
	ev.Status = payday.StatusDeferred
	pastDate := now.Add(-48 * time.Hour)
	ev.DeferUntil = &pastDate
	if err := f.paydayRepo.Create(f.ctx, ev); err != nil {
		t.Fatalf("create event: %v", err)
	}

	status, err := f.svc.GetCurrentMonthStatus(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if len(status.Portfolios) != 1 {
		t.Fatalf("expected 1 portfolio, got %d", len(status.Portfolios))
	}
	if status.Portfolios[0].Status != string(payday.StatusPending) {
		t.Errorf("status = %s, want %s", status.Portfolios[0].Status, payday.StatusPending)
	}
}

func TestGetCurrentMonthStatusDeferredFuture(t *testing.T) {
	f := setupPaydayTest(t)
	if err := f.svc.SavePaydayDay(f.ctx, 1); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	p := addTestPortfolio(t, f, "Deferred Future", 1000000)

	now := time.Now().UTC()
	currentMonth := now.Format("2006-01")

	ev := payday.NewPaydayEvent(currentMonth, p.ID, 1000000)
	ev.Status = payday.StatusDeferred
	futureDate := now.Add(72 * time.Hour)
	ev.DeferUntil = &futureDate
	if err := f.paydayRepo.Create(f.ctx, ev); err != nil {
		t.Fatalf("create event: %v", err)
	}

	status, err := f.svc.GetCurrentMonthStatus(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Portfolios[0].Status != string(payday.StatusDeferred) {
		t.Errorf("status = %s, want %s", status.Portfolios[0].Status, payday.StatusDeferred)
	}
}

func TestGetCurrentMonthStatusTotalExpected(t *testing.T) {
	f := setupPaydayTest(t)
	if err := f.svc.SavePaydayDay(f.ctx, 1); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	addTestPortfolio(t, f, "A", 3000000)
	addTestPortfolio(t, f, "B", 2000000)

	status, err := f.svc.GetCurrentMonthStatus(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.TotalExpected != 5000000 {
		t.Errorf("TotalExpected = %f, want %f", status.TotalExpected, 5000000.0)
	}
}

func TestConfirmPayday(t *testing.T) {
	t.Run("transitions PENDING to CONFIRMED and creates cash flow", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Confirm", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusPending)

		err := f.svc.ConfirmPayday(f.ctx, p.ID, 4500000)
		if err != nil {
			t.Fatalf("ConfirmPayday() error = %v", err)
		}

		// Verify event updated.
		ev, err := f.paydayRepo.GetByMonthAndPortfolio(f.ctx, currentMonth, p.ID)
		if err != nil {
			t.Fatalf("get event: %v", err)
		}
		if ev.Status != payday.StatusConfirmed {
			t.Errorf("status = %s, want %s", ev.Status, payday.StatusConfirmed)
		}
		if ev.Actual != 4500000 {
			t.Errorf("actual = %f, want %f", ev.Actual, 4500000.0)
		}
		if ev.ConfirmedAt == nil {
			t.Error("ConfirmedAt should not be nil")
		}

		// Verify cash flow created.
		flows, err := f.cashFlowRepo.ListByPortfolioID(f.ctx, p.ID)
		if err != nil {
			t.Fatalf("list cash flows: %v", err)
		}
		if len(flows) != 1 {
			t.Fatalf("expected 1 cash flow, got %d", len(flows))
		}
		if flows[0].Amount != 4500000 {
			t.Errorf("cash flow amount = %f, want %f", flows[0].Amount, 4500000.0)
		}
		if flows[0].Type != payday.FlowTypeMonthly {
			t.Errorf("cash flow type = %s, want %s", flows[0].Type, payday.FlowTypeMonthly)
		}
	})

	t.Run("rejects from SCHEDULED status", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Reject Scheduled", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusScheduled)

		err := f.svc.ConfirmPayday(f.ctx, p.ID, 5000000)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("rejects from CONFIRMED status", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Already Confirmed", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusConfirmed)

		err := f.svc.ConfirmPayday(f.ctx, p.ID, 5000000)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("rejects from SKIPPED status", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Skipped", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusSkipped)

		err := f.svc.ConfirmPayday(f.ctx, p.ID, 5000000)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})
}

func TestDeferPayday(t *testing.T) {
	t.Run("transitions PENDING to DEFERRED with defer date", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Defer", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusPending)

		deferDate := now.Add(7 * 24 * time.Hour)
		err := f.svc.DeferPayday(f.ctx, p.ID, deferDate)
		if err != nil {
			t.Fatalf("DeferPayday() error = %v", err)
		}

		ev, err := f.paydayRepo.GetByMonthAndPortfolio(f.ctx, currentMonth, p.ID)
		if err != nil {
			t.Fatalf("get event: %v", err)
		}
		if ev.Status != payday.StatusDeferred {
			t.Errorf("status = %s, want %s", ev.Status, payday.StatusDeferred)
		}
		if ev.DeferUntil == nil {
			t.Fatal("DeferUntil should not be nil")
		}
		if !ev.DeferUntil.Equal(deferDate) {
			t.Errorf("DeferUntil = %v, want %v", ev.DeferUntil, deferDate)
		}
	})

	t.Run("rejects past defer date", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Defer Past", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusPending)

		pastDate := now.Add(-24 * time.Hour)
		err := f.svc.DeferPayday(f.ctx, p.ID, pastDate)
		if !errors.Is(err, ErrDeferDateNotFuture) {
			t.Errorf("expected ErrDeferDateNotFuture, got %v", err)
		}
	})

	t.Run("rejects today as defer date", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Defer Today", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusPending)

		err := f.svc.DeferPayday(f.ctx, p.ID, now)
		if !errors.Is(err, ErrDeferDateNotFuture) {
			t.Errorf("expected ErrDeferDateNotFuture, got %v", err)
		}
	})

	t.Run("rejects from SCHEDULED status", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Defer Scheduled", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusScheduled)

		err := f.svc.DeferPayday(f.ctx, p.ID, now.Add(24*time.Hour))
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("rejects from CONFIRMED status", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Defer Confirmed", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusConfirmed)

		err := f.svc.DeferPayday(f.ctx, p.ID, now.Add(24*time.Hour))
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})
}

func TestSkipPayday(t *testing.T) {
	t.Run("transitions PENDING to SKIPPED", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Skip", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusPending)

		err := f.svc.SkipPayday(f.ctx, p.ID)
		if err != nil {
			t.Fatalf("SkipPayday() error = %v", err)
		}

		ev, err := f.paydayRepo.GetByMonthAndPortfolio(f.ctx, currentMonth, p.ID)
		if err != nil {
			t.Fatalf("get event: %v", err)
		}
		if ev.Status != payday.StatusSkipped {
			t.Errorf("status = %s, want %s", ev.Status, payday.StatusSkipped)
		}
	})

	t.Run("rejects from SCHEDULED status", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Skip Scheduled", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusScheduled)

		err := f.svc.SkipPayday(f.ctx, p.ID)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("rejects from CONFIRMED status", func(t *testing.T) {
		f := setupPaydayTest(t)
		p := addTestPortfolio(t, f, "Skip Confirmed", 5000000)

		now := time.Now().UTC()
		currentMonth := now.Format("2006-01")
		seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusConfirmed)

		err := f.svc.SkipPayday(f.ctx, p.ID)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})
}

func TestGetCashFlowSummaryEmpty(t *testing.T) {
	f := setupPaydayTest(t)
	p := addTestPortfolio(t, f, "Empty", 5000000)

	summary, err := f.svc.GetCashFlowSummary(f.ctx, p.ID)
	if err != nil {
		t.Fatalf("GetCashFlowSummary() error = %v", err)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if len(summary.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(summary.Items))
	}
	if summary.TotalInflow != 0 {
		t.Errorf("TotalInflow = %f, want 0", summary.TotalInflow)
	}
	if summary.Balance != 0 {
		t.Errorf("Balance = %f, want 0", summary.Balance)
	}
}

func TestGetCashFlowSummarySumsAmounts(t *testing.T) {
	f := setupPaydayTest(t)
	p := addTestPortfolio(t, f, "Flows", 5000000)

	now := time.Now().UTC()
	cf1 := payday.NewCashFlow(p.ID, payday.FlowTypeInitial, 10000000, now, "Initial deposit")
	cf2 := payday.NewCashFlow(p.ID, payday.FlowTypeMonthly, 5000000, now, "Monthly payday")
	cf3 := payday.NewCashFlow(p.ID, payday.FlowTypeDividend, 500000, now, "Dividend")

	for _, cf := range []*payday.CashFlow{cf1, cf2, cf3} {
		if err := f.cashFlowRepo.Create(f.ctx, cf); err != nil {
			t.Fatalf("create cash flow: %v", err)
		}
	}

	summary, err := f.svc.GetCashFlowSummary(f.ctx, p.ID)
	if err != nil {
		t.Fatalf("GetCashFlowSummary() error = %v", err)
	}
	if len(summary.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(summary.Items))
	}

	wantTotal := 15500000.0
	if summary.TotalInflow != wantTotal {
		t.Errorf("TotalInflow = %f, want %f", summary.TotalInflow, wantTotal)
	}
	if summary.Balance != wantTotal {
		t.Errorf("Balance = %f, want %f", summary.Balance, wantTotal)
	}
}

func TestGetCashFlowSummaryItemFields(t *testing.T) {
	f := setupPaydayTest(t)
	p := addTestPortfolio(t, f, "Fields", 1000000)

	now := time.Now().UTC()
	cf := payday.NewCashFlow(p.ID, payday.FlowTypeMonthly, 1000000, now, "Test note")
	if err := f.cashFlowRepo.Create(f.ctx, cf); err != nil {
		t.Fatalf("create cash flow: %v", err)
	}

	summary, err := f.svc.GetCashFlowSummary(f.ctx, p.ID)
	if err != nil {
		t.Fatalf("GetCashFlowSummary() error = %v", err)
	}
	if len(summary.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(summary.Items))
	}
	item := summary.Items[0]
	if item.PortfolioID != p.ID {
		t.Errorf("PortfolioID = %s, want %s", item.PortfolioID, p.ID)
	}
	if item.Type != string(payday.FlowTypeMonthly) {
		t.Errorf("Type = %s, want %s", item.Type, payday.FlowTypeMonthly)
	}
	if item.Amount != 1000000 {
		t.Errorf("Amount = %f, want %f", item.Amount, 1000000.0)
	}
	if item.Note != "Test note" {
		t.Errorf("Note = %s, want %s", item.Note, "Test note")
	}
}

func TestGetCashFlowSummaryTotalDeployed(t *testing.T) {
	tests := []struct {
		name              string
		cashFlows         []*payday.CashFlow
		txnRecords        []transaction.Record
		wantTotalInflow   float64
		wantTotalDeployed float64
		wantBalance       float64
	}{
		{
			name:              "no transactions yields zero totalDeployed",
			wantTotalInflow:   0,
			wantTotalDeployed: 0,
			wantBalance:       0,
		},
		{
			name: "totalDeployed equals TotalBuyAmount from transactions",
			cashFlows: []*payday.CashFlow{
				payday.NewCashFlow("p1", payday.FlowTypeMonthly, 10000000, time.Now().UTC(), "Monthly"),
			},
			txnRecords: []transaction.Record{
				{PortfolioID: "p1", Type: transaction.TypeBuy, Total: 3000000, Fee: 15000},
				{PortfolioID: "p1", Type: transaction.TypeBuy, Total: 2000000, Fee: 10000},
			},
			wantTotalInflow:   10000000,
			wantTotalDeployed: 5000000,
			wantBalance:       5000000,
		},
		{
			name: "sell transactions do not affect totalDeployed",
			cashFlows: []*payday.CashFlow{
				payday.NewCashFlow("p1", payday.FlowTypeMonthly, 10000000, time.Now().UTC(), "Monthly"),
			},
			txnRecords: []transaction.Record{
				{PortfolioID: "p1", Type: transaction.TypeBuy, Total: 4000000, Fee: 20000},
				{PortfolioID: "p1", Type: transaction.TypeSell, Total: 1000000, Fee: 5000},
			},
			wantTotalInflow:   10000000,
			wantTotalDeployed: 4000000,
			wantBalance:       6000000,
		},
		{
			name: "filters transactions by portfolio ID",
			cashFlows: []*payday.CashFlow{
				payday.NewCashFlow("p1", payday.FlowTypeMonthly, 10000000, time.Now().UTC(), "Monthly"),
			},
			txnRecords: []transaction.Record{
				{PortfolioID: "p1", Type: transaction.TypeBuy, Total: 3000000, Fee: 15000},
				{PortfolioID: "other", Type: transaction.TypeBuy, Total: 7000000, Fee: 35000},
			},
			wantTotalInflow:   10000000,
			wantTotalDeployed: 3000000,
			wantBalance:       7000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupPaydayTest(t)

			portfolioID := "p1"
			for _, cf := range tt.cashFlows {
				if err := f.cashFlowRepo.Create(f.ctx, cf); err != nil {
					t.Fatalf("create cash flow: %v", err)
				}
			}
			f.txnRepo.mu.Lock()
			f.txnRepo.records = tt.txnRecords
			f.txnRepo.mu.Unlock()

			summary, err := f.svc.GetCashFlowSummary(f.ctx, portfolioID)
			if err != nil {
				t.Fatalf("GetCashFlowSummary() error = %v", err)
			}
			if summary.TotalInflow != tt.wantTotalInflow {
				t.Errorf("TotalInflow = %f, want %f", summary.TotalInflow, tt.wantTotalInflow)
			}
			if summary.TotalDeployed != tt.wantTotalDeployed {
				t.Errorf("TotalDeployed = %f, want %f", summary.TotalDeployed, tt.wantTotalDeployed)
			}
			if summary.Balance != tt.wantBalance {
				t.Errorf("Balance = %f, want %f", summary.Balance, tt.wantBalance)
			}
		})
	}
}

func TestGetPaydayHistoryEmpty(t *testing.T) {
	f := setupPaydayTest(t)
	if err := f.svc.SavePaydayDay(f.ctx, 25); err != nil {
		t.Fatalf("save payday day: %v", err)
	}

	history, err := f.svc.GetPaydayHistory(f.ctx)
	if err != nil {
		t.Fatalf("GetPaydayHistory() error = %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected 0 months, got %d", len(history))
	}
}

func TestGetPaydayHistoryExcludesCurrentMonth(t *testing.T) {
	f := setupPaydayTest(t)
	if err := f.svc.SavePaydayDay(f.ctx, 25); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	p := addTestPortfolio(t, f, "History", 5000000)

	now := time.Now().UTC()
	currentMonth := now.Format("2006-01")

	// Seed event in current month — should be excluded from history.
	seedPaydayEvent(t, f, currentMonth, p.ID, payday.StatusPending)

	history, err := f.svc.GetPaydayHistory(f.ctx)
	if err != nil {
		t.Fatalf("GetPaydayHistory() error = %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected 0 months (current excluded), got %d", len(history))
	}
}

func TestGetPaydayHistoryReturnsPastMonths(t *testing.T) {
	f := setupPaydayTest(t)
	if err := f.svc.SavePaydayDay(f.ctx, 25); err != nil {
		t.Fatalf("save payday day: %v", err)
	}
	p := addTestPortfolio(t, f, "History", 5000000)

	// Seed events in past months.
	seedPaydayEvent(t, f, "2025-12", p.ID, payday.StatusConfirmed)
	seedPaydayEvent(t, f, "2026-01", p.ID, payday.StatusSkipped)

	history, err := f.svc.GetPaydayHistory(f.ctx)
	if err != nil {
		t.Fatalf("GetPaydayHistory() error = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 months, got %d", len(history))
	}
	// Should be sorted descending.
	if history[0].Month != "2026-01" {
		t.Errorf("first month = %s, want 2026-01", history[0].Month)
	}
	if history[1].Month != "2025-12" {
		t.Errorf("second month = %s, want 2025-12", history[1].Month)
	}
	if len(history[0].Portfolios) != 1 {
		t.Errorf("expected 1 portfolio in 2026-01, got %d", len(history[0].Portfolios))
	}
}
