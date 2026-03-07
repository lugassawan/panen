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

type sellTransactionTestFixture struct {
	repo    *SellTransactionRepo
	holding *portfolio.Holding
	ctx     context.Context
	now     time.Time
}

func setupSellTransactionTest(t *testing.T) sellTransactionTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	broRepo := NewBrokerageRepo(db)
	portRepo := NewPortfolioRepo(db)
	holdRepo := NewHoldingRepo(db)

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
	h := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: port.ID, Ticker: "BBCA",
		AvgBuyPrice: 8500, Lots: 10,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := holdRepo.Create(ctx, h); err != nil {
		t.Fatalf("create holding: %v", err)
	}

	return sellTransactionTestFixture{
		repo:    NewSellTransactionRepo(db),
		holding: h,
		ctx:     ctx,
		now:     now,
	}
}

func TestSellTransactionRepoCreateAndGetByID(t *testing.T) {
	f := setupSellTransactionTest(t)

	tx := &portfolio.SellTransaction{
		ID: shared.NewID(), HoldingID: f.holding.ID,
		Date: f.now, Price: 9000, Lots: 3, Fee: 6750, Tax: 2700,
		RealizedGain: 150000,
		CreatedAt:    f.now,
	}
	if err := f.repo.Create(f.ctx, tx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Price != 9000 {
		t.Errorf("Price = %f, want 9000", got.Price)
	}
	if got.Lots != 3 {
		t.Errorf("Lots = %d, want 3", got.Lots)
	}
	if got.Fee != 6750 {
		t.Errorf("Fee = %f, want 6750", got.Fee)
	}
	if got.Tax != 2700 {
		t.Errorf("Tax = %f, want 2700", got.Tax)
	}
	if got.RealizedGain != 150000 {
		t.Errorf("RealizedGain = %f, want 150000", got.RealizedGain)
	}
}

func TestSellTransactionRepoGetByIDNotFound(t *testing.T) {
	f := setupSellTransactionTest(t)

	_, err := f.repo.GetByID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestSellTransactionRepoListByHoldingID(t *testing.T) {
	f := setupSellTransactionTest(t)

	tx := &portfolio.SellTransaction{
		ID: shared.NewID(), HoldingID: f.holding.ID,
		Date: f.now, Price: 9000, Lots: 2, Fee: 4500, Tax: 1800,
		RealizedGain: 100000,
		CreatedAt:    f.now,
	}
	if err := f.repo.Create(f.ctx, tx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := f.repo.ListByHoldingID(f.ctx, f.holding.ID)
	if err != nil {
		t.Fatalf("ListByHoldingID() error = %v", err)
	}
	if len(list) < 1 {
		t.Error("ListByHoldingID() returned empty slice")
	}
}

func TestSellTransactionRepoDelete(t *testing.T) {
	f := setupSellTransactionTest(t)

	tx := &portfolio.SellTransaction{
		ID: shared.NewID(), HoldingID: f.holding.ID,
		Date: f.now, Price: 9500, Lots: 1, Fee: 2375, Tax: 950,
		RealizedGain: 100000,
		CreatedAt:    f.now,
	}
	if err := f.repo.Create(f.ctx, tx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := f.repo.Delete(f.ctx, tx.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, err := f.repo.GetByID(f.ctx, tx.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after Delete error = %v, want ErrNotFound", err)
	}
}
