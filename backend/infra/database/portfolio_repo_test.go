package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
)

type portfolioTestFixture struct {
	repo     *PortfolioRepo
	broRepo  *BrokerageRepo
	userRepo *UserRepo
	account  *brokerage.Account
	ctx      context.Context
	now      time.Time
}

func setupPortfolioTest(t *testing.T) portfolioTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	broRepo := NewBrokerageRepo(db)

	p := &user.Profile{
		ID: shared.NewID(), Name: "Test User",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := userRepo.Create(ctx, p); err != nil {
		t.Fatalf("create profile: %v", err)
	}
	a := &brokerage.Account{
		ID: shared.NewID(), ProfileID: p.ID, BrokerName: "Broker",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := broRepo.Create(ctx, a); err != nil {
		t.Fatalf("create brokerage: %v", err)
	}

	return portfolioTestFixture{
		repo:     NewPortfolioRepo(db),
		broRepo:  broRepo,
		userRepo: userRepo,
		account:  a,
		ctx:      ctx,
		now:      now,
	}
}

func TestPortfolioRepoCreateAndGetByID(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.account.ID,
		Name:               "Growth",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileAggressive,
		Capital:            10000000,
		MonthlyAddition:    1000000,
		MaxStocks:          10,
		Universe:           []string{"BBCA", "BBRI", "BMRI"},
		CreatedAt:          f.now,
		UpdatedAt:          f.now,
	}
	if err := f.repo.Create(f.ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Growth" {
		t.Errorf("Name = %q, want %q", got.Name, "Growth")
	}
	if got.Mode != portfolio.ModeValue {
		t.Errorf("Mode = %q, want %q", got.Mode, portfolio.ModeValue)
	}
	if len(got.Universe) != 3 || got.Universe[0] != "BBCA" {
		t.Errorf("Universe = %v, want [BBCA BBRI BMRI]", got.Universe)
	}
}

func TestPortfolioRepoEmptyUniverseRoundtrip(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.account.ID,
		Name:               "Empty",
		Mode:               portfolio.ModeDividend,
		RiskProfile:        portfolio.RiskProfileConservative,
		Universe:           []string{},
		CreatedAt:          f.now,
		UpdatedAt:          f.now,
	}
	if err := f.repo.Create(f.ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Universe == nil {
		t.Error("Universe is nil, want empty slice")
	}
	if len(got.Universe) != 0 {
		t.Errorf("Universe len = %d, want 0", len(got.Universe))
	}
}

func TestPortfolioRepoGetByIDNotFound(t *testing.T) {
	f := setupPortfolioTest(t)

	_, err := f.repo.GetByID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestPortfolioRepoListByBrokerageAccountID(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.account.ID,
		Name:               "List Test",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          f.now,
		UpdatedAt:          f.now,
	}
	if err := f.repo.Create(f.ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := f.repo.ListByBrokerageAccountID(f.ctx, f.account.ID)
	if err != nil {
		t.Fatalf("ListByBrokerageAccountID() error = %v", err)
	}
	if len(list) < 1 {
		t.Error("ListByBrokerageAccountID() returned empty slice")
	}
}

func TestPortfolioRepoUpdate(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.account.ID,
		Name:               "Old Name",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          f.now,
		UpdatedAt:          f.now,
	}
	if err := f.repo.Create(f.ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	p.Name = "New Name"
	p.Universe = []string{"TLKM"}
	p.UpdatedAt = f.now.Add(time.Hour)
	if err := f.repo.Update(f.ctx, p); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "New Name" {
		t.Errorf("Name = %q, want %q", got.Name, "New Name")
	}
	if len(got.Universe) != 1 || got.Universe[0] != "TLKM" {
		t.Errorf("Universe = %v, want [TLKM]", got.Universe)
	}
}

func TestPortfolioRepoCascadeDeleteFromBrokerage(t *testing.T) {
	f := setupPortfolioTest(t)

	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: f.account.ID,
		Name:               "Cascade",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          f.now,
		UpdatedAt:          f.now,
	}
	if err := f.repo.Create(f.ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := f.broRepo.Delete(f.ctx, f.account.ID); err != nil {
		t.Fatalf("Delete brokerage error = %v", err)
	}

	_, err := f.repo.GetByID(f.ctx, p.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after cascade error = %v, want ErrNotFound", err)
	}
}
