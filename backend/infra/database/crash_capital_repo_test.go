package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/crashplaybook"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
)

type crashCapitalTestFixture struct {
	repo      *CrashCapitalRepo
	portfolio *portfolio.Portfolio
	ctx       context.Context
	now       time.Time
}

func setupCrashCapitalTest(t *testing.T) crashCapitalTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	broRepo := NewBrokerageRepo(db)
	portRepo := NewPortfolioRepo(db)

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
	port := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: a.ID,
		Name:               "Test Portfolio",
		Mode:               portfolio.ModeValue,
		RiskProfile:        portfolio.RiskProfileConservative,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := portRepo.Create(ctx, port); err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	return crashCapitalTestFixture{
		repo:      NewCrashCapitalRepo(db),
		portfolio: port,
		ctx:       ctx,
		now:       now,
	}
}

func TestCrashCapitalRepoUpsertAndGet(t *testing.T) {
	f := setupCrashCapitalTest(t)

	cc := &crashplaybook.CrashCapital{
		ID:          shared.NewID(),
		PortfolioID: f.portfolio.ID,
		Amount:      10_000_000,
		Deployed:    0,
		CreatedAt:   f.now,
		UpdatedAt:   f.now,
	}
	if err := f.repo.Upsert(f.ctx, cc); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := f.repo.GetByPortfolioID(f.ctx, f.portfolio.ID)
	if err != nil {
		t.Fatalf("GetByPortfolioID() error = %v", err)
	}
	if got.Amount != 10_000_000 {
		t.Errorf("Amount = %v, want 10000000", got.Amount)
	}
	if got.PortfolioID != f.portfolio.ID {
		t.Errorf("PortfolioID = %q, want %q", got.PortfolioID, f.portfolio.ID)
	}
}

func TestCrashCapitalRepoUpsertUpdatesExisting(t *testing.T) {
	f := setupCrashCapitalTest(t)

	cc := &crashplaybook.CrashCapital{
		ID:          shared.NewID(),
		PortfolioID: f.portfolio.ID,
		Amount:      10_000_000,
		CreatedAt:   f.now,
		UpdatedAt:   f.now,
	}
	if err := f.repo.Upsert(f.ctx, cc); err != nil {
		t.Fatalf("Upsert() insert error = %v", err)
	}

	cc.Amount = 20_000_000
	cc.UpdatedAt = f.now.Add(time.Hour)
	if err := f.repo.Upsert(f.ctx, cc); err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}

	got, err := f.repo.GetByPortfolioID(f.ctx, f.portfolio.ID)
	if err != nil {
		t.Fatalf("GetByPortfolioID() error = %v", err)
	}
	if got.Amount != 20_000_000 {
		t.Errorf("Amount = %v, want 20000000", got.Amount)
	}
}

func TestCrashCapitalRepoGetByPortfolioIDNotFound(t *testing.T) {
	f := setupCrashCapitalTest(t)

	_, err := f.repo.GetByPortfolioID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByPortfolioID() error = %v, want ErrNotFound", err)
	}
}

func TestCrashCapitalRepoTimestampRoundTrip(t *testing.T) {
	f := setupCrashCapitalTest(t)

	cc := &crashplaybook.CrashCapital{
		ID:          shared.NewID(),
		PortfolioID: f.portfolio.ID,
		Amount:      5_000_000,
		CreatedAt:   f.now,
		UpdatedAt:   f.now.Add(2 * time.Hour),
	}
	if err := f.repo.Upsert(f.ctx, cc); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := f.repo.GetByPortfolioID(f.ctx, f.portfolio.ID)
	if err != nil {
		t.Fatalf("GetByPortfolioID() error = %v", err)
	}
	if !got.CreatedAt.Equal(f.now) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, f.now)
	}
	if !got.UpdatedAt.Equal(f.now.Add(2 * time.Hour)) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, f.now.Add(2*time.Hour))
	}
}
