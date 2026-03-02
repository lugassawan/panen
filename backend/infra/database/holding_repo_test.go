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

type holdingTestFixture struct {
	repo          *HoldingRepo
	portfolioRepo *PortfolioRepo
	port          *portfolio.Portfolio
	ctx           context.Context
	now           time.Time
}

func setupHoldingTest(t *testing.T) holdingTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	broRepo := NewBrokerageRepo(db)
	portRepo := NewPortfolioRepo(db)

	p := &user.Profile{
		ID: shared.NewID(), Name: "User",
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
	port := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: a.ID,
		Name:               "Test",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileModerate,
		Universe:           []string{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := portRepo.Create(ctx, port); err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	return holdingTestFixture{
		repo:          NewHoldingRepo(db),
		portfolioRepo: portRepo,
		port:          port,
		ctx:           ctx,
		now:           now,
	}
}

func TestHoldingRepoCreateAndGetByID(t *testing.T) {
	f := setupHoldingTest(t)

	h := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID,
		Ticker: "BBCA", AvgBuyPrice: 8500, Lots: 10,
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, h); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, h.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want %q", got.Ticker, "BBCA")
	}
	if got.Lots != 10 {
		t.Errorf("Lots = %d, want 10", got.Lots)
	}
}

func TestHoldingRepoGetByIDNotFound(t *testing.T) {
	f := setupHoldingTest(t)

	_, err := f.repo.GetByID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestHoldingRepoUniqueConstraint(t *testing.T) {
	f := setupHoldingTest(t)

	h1 := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID, Ticker: "UNIQ",
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, h1); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	h2 := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID, Ticker: "UNIQ",
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, h2); err == nil {
		t.Fatal("Create() duplicate should fail")
	}
}

func TestHoldingRepoGetByPortfolioAndTicker(t *testing.T) {
	f := setupHoldingTest(t)

	h := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID,
		Ticker: "BBCA", AvgBuyPrice: 8500, Lots: 10,
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, h); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByPortfolioAndTicker(f.ctx, f.port.ID, "BBCA")
	if err != nil {
		t.Fatalf("GetByPortfolioAndTicker() error = %v", err)
	}
	if got.ID != h.ID {
		t.Errorf("ID = %q, want %q", got.ID, h.ID)
	}
	if got.AvgBuyPrice != 8500 {
		t.Errorf("AvgBuyPrice = %f, want 8500", got.AvgBuyPrice)
	}
}

func TestHoldingRepoGetByPortfolioAndTickerNotFound(t *testing.T) {
	f := setupHoldingTest(t)

	_, err := f.repo.GetByPortfolioAndTicker(f.ctx, f.port.ID, "NONEXIST")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByPortfolioAndTicker() error = %v, want ErrNotFound", err)
	}
}

func TestHoldingRepoListByPortfolioID(t *testing.T) {
	f := setupHoldingTest(t)

	h := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID, Ticker: "BBCA",
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, h); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := f.repo.ListByPortfolioID(f.ctx, f.port.ID)
	if err != nil {
		t.Fatalf("ListByPortfolioID() error = %v", err)
	}
	if len(list) < 1 {
		t.Error("ListByPortfolioID() returned empty slice")
	}
}

func TestHoldingRepoUpdate(t *testing.T) {
	f := setupHoldingTest(t)

	h := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID, Ticker: "BMRI",
		AvgBuyPrice: 5000, Lots: 5,
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, h); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	h.AvgBuyPrice = 5500
	h.Lots = 15
	h.UpdatedAt = f.now.Add(time.Hour)
	if err := f.repo.Update(f.ctx, h); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, h.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.AvgBuyPrice != 5500 {
		t.Errorf("AvgBuyPrice = %f, want 5500", got.AvgBuyPrice)
	}
	if got.Lots != 15 {
		t.Errorf("Lots = %d, want 15", got.Lots)
	}
}

func TestHoldingRepoCascadeDeleteFromPortfolio(t *testing.T) {
	f := setupHoldingTest(t)

	h := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID, Ticker: "TLKM",
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, h); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := f.portfolioRepo.Delete(f.ctx, f.port.ID); err != nil {
		t.Fatalf("Delete portfolio error = %v", err)
	}

	_, err := f.repo.GetByID(f.ctx, h.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after cascade error = %v, want ErrNotFound", err)
	}
}
