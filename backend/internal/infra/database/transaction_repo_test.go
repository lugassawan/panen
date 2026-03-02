package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/brokerage"
	"github.com/lugassawan/panen/backend/internal/domain/portfolio"
	"github.com/lugassawan/panen/backend/internal/domain/shared"
	"github.com/lugassawan/panen/backend/internal/domain/user"
)

type transactionTestFixture struct {
	repo        *BuyTransactionRepo
	holdingRepo *HoldingRepo
	holding     *portfolio.Holding
	port        *portfolio.Portfolio
	ctx         context.Context
	now         time.Time
}

func setupTransactionTest(t *testing.T) transactionTestFixture {
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

	return transactionTestFixture{
		repo:        NewBuyTransactionRepo(db),
		holdingRepo: holdRepo,
		holding:     h,
		port:        port,
		ctx:         ctx,
		now:         now,
	}
}

func TestBuyTransactionRepoCreateAndGetByID(t *testing.T) {
	f := setupTransactionTest(t)

	tx := &portfolio.BuyTransaction{
		ID: shared.NewID(), HoldingID: f.holding.ID,
		Date: f.now, Price: 8500, Lots: 5, Fee: 63750,
		CreatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, tx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Price != 8500 {
		t.Errorf("Price = %f, want 8500", got.Price)
	}
	if got.Lots != 5 {
		t.Errorf("Lots = %d, want 5", got.Lots)
	}
	if got.Fee != 63750 {
		t.Errorf("Fee = %f, want 63750", got.Fee)
	}
}

func TestBuyTransactionRepoGetByIDNotFound(t *testing.T) {
	f := setupTransactionTest(t)

	_, err := f.repo.GetByID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestBuyTransactionRepoListByHoldingID(t *testing.T) {
	f := setupTransactionTest(t)

	tx := &portfolio.BuyTransaction{
		ID: shared.NewID(), HoldingID: f.holding.ID,
		Date: f.now, Price: 8500, Lots: 5, CreatedAt: f.now,
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

func TestBuyTransactionRepoDelete(t *testing.T) {
	f := setupTransactionTest(t)

	tx := &portfolio.BuyTransaction{
		ID: shared.NewID(), HoldingID: f.holding.ID,
		Date: f.now, Price: 8600, Lots: 2, Fee: 25800,
		CreatedAt: f.now,
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

func TestBuyTransactionRepoCascadeDeleteFromHolding(t *testing.T) {
	f := setupTransactionTest(t)

	// Create a separate holding for cascade test.
	h := &portfolio.Holding{
		ID: shared.NewID(), PortfolioID: f.port.ID, Ticker: "BBRI",
		CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.holdingRepo.Create(f.ctx, h); err != nil {
		t.Fatalf("create holding: %v", err)
	}
	tx := &portfolio.BuyTransaction{
		ID: shared.NewID(), HoldingID: h.ID,
		Date: f.now, Price: 4500, Lots: 3, Fee: 20250,
		CreatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, tx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := f.holdingRepo.Delete(f.ctx, h.ID); err != nil {
		t.Fatalf("Delete holding error = %v", err)
	}

	_, err := f.repo.GetByID(f.ctx, tx.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after cascade error = %v, want ErrNotFound", err)
	}
}
