package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
)

type brokerageTestFixture struct {
	userRepo *UserRepo
	repo     *BrokerageRepo
	profile  *user.Profile
	ctx      context.Context
	now      time.Time
}

func setupBrokerageTest(t *testing.T) brokerageTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	p := &user.Profile{
		ID: shared.NewID(), Name: "Test User",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := userRepo.Create(ctx, p); err != nil {
		t.Fatalf("create test profile: %v", err)
	}

	return brokerageTestFixture{
		userRepo: userRepo,
		repo:     NewBrokerageRepo(db),
		profile:  p,
		ctx:      ctx,
		now:      now,
	}
}

func TestBrokerageRepoCreateAndGetByID(t *testing.T) {
	f := setupBrokerageTest(t)

	a := &brokerage.Account{
		ID: shared.NewID(), ProfileID: f.profile.ID,
		BrokerName: "Ajaib", BrokerCode: "AJAIB", BuyFeePct: 0.15, SellFeePct: 0.25,
		SellTaxPct: 0.1, IsManualFee: true, CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, a.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.BrokerName != "Ajaib" {
		t.Errorf("BrokerName = %q, want %q", got.BrokerName, "Ajaib")
	}
	if got.BrokerCode != "AJAIB" {
		t.Errorf("BrokerCode = %q, want %q", got.BrokerCode, "AJAIB")
	}
	if !got.IsManualFee {
		t.Error("IsManualFee = false, want true")
	}
	if got.BuyFeePct != 0.15 {
		t.Errorf("BuyFeePct = %f, want 0.15", got.BuyFeePct)
	}
	if got.SellTaxPct != 0.1 {
		t.Errorf("SellTaxPct = %f, want 0.1", got.SellTaxPct)
	}
}

func TestBrokerageRepoGetByIDNotFound(t *testing.T) {
	f := setupBrokerageTest(t)

	_, err := f.repo.GetByID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestBrokerageRepoListByProfileID(t *testing.T) {
	f := setupBrokerageTest(t)

	a := &brokerage.Account{
		ID: shared.NewID(), ProfileID: f.profile.ID,
		BrokerName: "Broker", CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	accounts, err := f.repo.ListByProfileID(f.ctx, f.profile.ID)
	if err != nil {
		t.Fatalf("ListByProfileID() error = %v", err)
	}
	if len(accounts) < 1 {
		t.Error("ListByProfileID() returned empty slice")
	}
}

func TestBrokerageRepoUpdate(t *testing.T) {
	f := setupBrokerageTest(t)

	a := &brokerage.Account{
		ID: shared.NewID(), ProfileID: f.profile.ID,
		BrokerName: "Stockbit", CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	a.BrokerName = "Bibit"
	a.UpdatedAt = f.now.Add(time.Hour)
	if err := f.repo.Update(f.ctx, a); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, a.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.BrokerName != "Bibit" {
		t.Errorf("BrokerName = %q, want %q", got.BrokerName, "Bibit")
	}
}

func TestBrokerageRepoCascadeDeleteFromProfile(t *testing.T) {
	f := setupBrokerageTest(t)

	a := &brokerage.Account{
		ID: shared.NewID(), ProfileID: f.profile.ID,
		BrokerName: "IPOT", CreatedAt: f.now, UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := f.userRepo.Delete(f.ctx, f.profile.ID); err != nil {
		t.Fatalf("Delete profile error = %v", err)
	}

	_, err := f.repo.GetByID(f.ctx, a.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after cascade error = %v, want ErrNotFound", err)
	}
}
