package database

import (
	"errors"
	"testing"

	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
)

func TestCashFlowRepoCreateAndList(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeValue)
	repo := NewCashFlowRepo(f.DB)

	cf := &payday.CashFlow{
		ID: shared.NewID(), PortfolioID: f.PortfolioID,
		Type: payday.FlowTypeMonthly, Amount: 500000,
		Date: f.Now, Note: "Monthly deposit", CreatedAt: f.Now,
	}
	if err := repo.Create(f.Ctx, cf); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	flows, err := repo.ListByPortfolioID(f.Ctx, f.PortfolioID)
	if err != nil {
		t.Fatalf("ListByPortfolioID() error = %v", err)
	}
	if len(flows) != 1 {
		t.Fatalf("got %d flows, want 1", len(flows))
	}

	got := flows[0]
	if got.ID != cf.ID {
		t.Errorf("ID = %q, want %q", got.ID, cf.ID)
	}
	if got.Type != payday.FlowTypeMonthly {
		t.Errorf("Type = %q, want %q", got.Type, payday.FlowTypeMonthly)
	}
	if got.Amount != 500000 {
		t.Errorf("Amount = %v, want 500000", got.Amount)
	}
	if got.Note != "Monthly deposit" {
		t.Errorf("Note = %q, want %q", got.Note, "Monthly deposit")
	}
}

func TestCashFlowRepoDelete(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeValue)
	repo := NewCashFlowRepo(f.DB)

	cf := &payday.CashFlow{
		ID: shared.NewID(), PortfolioID: f.PortfolioID,
		Type: payday.FlowTypeDividend, Amount: 100000,
		Date: f.Now, Note: "Dividend", CreatedAt: f.Now,
	}
	if err := repo.Create(f.Ctx, cf); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(f.Ctx, cf.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	flows, err := repo.ListByPortfolioID(f.Ctx, f.PortfolioID)
	if err != nil {
		t.Fatalf("ListByPortfolioID() error = %v", err)
	}
	if len(flows) != 0 {
		t.Errorf("got %d flows after delete, want 0", len(flows))
	}
}

func TestCashFlowRepoDeleteNotFound(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeValue)
	repo := NewCashFlowRepo(f.DB)

	err := repo.Delete(f.Ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestCashFlowRepoListEmptyPortfolio(t *testing.T) {
	f := setupDBFixture(t, portfolio.ModeValue)
	repo := NewCashFlowRepo(f.DB)

	flows, err := repo.ListByPortfolioID(f.Ctx, f.PortfolioID)
	if err != nil {
		t.Fatalf("ListByPortfolioID() error = %v", err)
	}
	if len(flows) != 0 {
		t.Errorf("got %d flows, want 0", len(flows))
	}
}
