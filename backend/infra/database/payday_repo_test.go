package database

import (
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
)

func TestPaydayRepoCreateAndGetByMonthAndPortfolio(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeDividend)
	repo := NewPaydayRepo(f.DB)

	event := &payday.PaydayEvent{
		ID: shared.NewID(), Month: "2025-06", PortfolioID: f.PortfolioID,
		Expected: 1000000, Actual: 0, Status: payday.StatusScheduled,
		CreatedAt: f.Now, UpdatedAt: f.Now,
	}
	if err := repo.Create(f.Ctx, event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByMonthAndPortfolio(f.Ctx, "2025-06", f.PortfolioID)
	if err != nil {
		t.Fatalf("GetByMonthAndPortfolio() error = %v", err)
	}

	if got.ID != event.ID {
		t.Errorf("ID = %q, want %q", got.ID, event.ID)
	}
	if got.Month != "2025-06" {
		t.Errorf("Month = %q, want %q", got.Month, "2025-06")
	}
	if got.Expected != 1000000 {
		t.Errorf("Expected = %v, want 1000000", got.Expected)
	}
	if got.Status != payday.StatusScheduled {
		t.Errorf("Status = %q, want %q", got.Status, payday.StatusScheduled)
	}
}

func TestPaydayRepoGetByMonthAndPortfolioNotFound(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeDividend)
	repo := NewPaydayRepo(f.DB)

	_, err := repo.GetByMonthAndPortfolio(f.Ctx, "2025-06", f.PortfolioID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPaydayRepoListByMonth(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeDividend)
	repo := NewPaydayRepo(f.DB)

	event := &payday.PaydayEvent{
		ID: shared.NewID(), Month: "2025-07", PortfolioID: f.PortfolioID,
		Expected: 500000, Status: payday.StatusScheduled,
		CreatedAt: f.Now, UpdatedAt: f.Now,
	}
	if err := repo.Create(f.Ctx, event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	events, err := repo.ListByMonth(f.Ctx, "2025-07")
	if err != nil {
		t.Fatalf("ListByMonth() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].ID != event.ID {
		t.Errorf("ID = %q, want %q", events[0].ID, event.ID)
	}
}

func TestPaydayRepoNullableTimeFields(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeDividend)
	repo := NewPaydayRepo(f.DB)

	deferTime := f.Now.Add(24 * time.Hour)
	event := &payday.PaydayEvent{
		ID: shared.NewID(), Month: "2025-08", PortfolioID: f.PortfolioID,
		Expected: 300000, Status: payday.StatusDeferred,
		DeferUntil: &deferTime,
		CreatedAt:  f.Now, UpdatedAt: f.Now,
	}
	if err := repo.Create(f.Ctx, event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByMonthAndPortfolio(f.Ctx, "2025-08", f.PortfolioID)
	if err != nil {
		t.Fatalf("GetByMonthAndPortfolio() error = %v", err)
	}

	if got.DeferUntil == nil {
		t.Fatal("expected DeferUntil to be set")
	}
	if !got.DeferUntil.Equal(deferTime) {
		t.Errorf("DeferUntil = %v, want %v", *got.DeferUntil, deferTime)
	}
	if got.ConfirmedAt != nil {
		t.Errorf("expected ConfirmedAt to be nil, got %v", got.ConfirmedAt)
	}
}

func TestPaydayRepoUpdate(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeDividend)
	repo := NewPaydayRepo(f.DB)

	event := &payday.PaydayEvent{
		ID: shared.NewID(), Month: "2025-09", PortfolioID: f.PortfolioID,
		Expected: 200000, Status: payday.StatusScheduled,
		CreatedAt: f.Now, UpdatedAt: f.Now,
	}
	if err := repo.Create(f.Ctx, event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	confirmedAt := f.Now.Add(2 * time.Hour)
	event.Actual = 250000
	event.Status = payday.StatusConfirmed
	event.ConfirmedAt = &confirmedAt
	event.UpdatedAt = f.Now.Add(2 * time.Hour)

	if err := repo.Update(f.Ctx, event); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetByMonthAndPortfolio(f.Ctx, "2025-09", f.PortfolioID)
	if err != nil {
		t.Fatalf("GetByMonthAndPortfolio() error = %v", err)
	}

	if got.Actual != 250000 {
		t.Errorf("Actual = %v, want 250000", got.Actual)
	}
	if got.Status != payday.StatusConfirmed {
		t.Errorf("Status = %q, want %q", got.Status, payday.StatusConfirmed)
	}
	if got.ConfirmedAt == nil {
		t.Fatal("expected ConfirmedAt to be set after confirm")
	}
}

func TestPaydayRepoUpdateNotFound(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeDividend)
	repo := NewPaydayRepo(f.DB)

	event := &payday.PaydayEvent{
		ID: "nonexistent", Month: "2025-10", PortfolioID: f.PortfolioID,
		Status: payday.StatusScheduled, UpdatedAt: f.Now,
	}
	err := repo.Update(f.Ctx, event)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Update() error = %v, want ErrNotFound", err)
	}
}
