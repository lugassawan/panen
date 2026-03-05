package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/trailingstop"
	"github.com/lugassawan/panen/backend/domain/user"
)

type peakRepoTestFixture struct {
	repo     *PeakRepo
	holdings []*portfolio.Holding
	ctx      context.Context
	now      time.Time
}

func setupPeakRepoTest(t *testing.T) peakRepoTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)
	broRepo := NewBrokerageRepo(db)
	portRepo := NewPortfolioRepo(db)
	holdRepo := NewHoldingRepo(db)

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

	holdings := make([]*portfolio.Holding, 0, 2)
	for _, ticker := range []string{"BBCA", "TLKM"} {
		h := portfolio.NewHolding(port.ID, ticker, 8000, 10)
		if err := holdRepo.Create(ctx, h); err != nil {
			t.Fatalf("create holding %s: %v", ticker, err)
		}
		holdings = append(holdings, h)
	}

	return peakRepoTestFixture{
		repo:     NewPeakRepo(db),
		holdings: holdings,
		ctx:      ctx,
		now:      now,
	}
}

func TestPeakRepoUpsertAndGet(t *testing.T) {
	f := setupPeakRepoTest(t)

	peak := &trailingstop.HoldingPeak{
		ID:        shared.NewID(),
		HoldingID: f.holdings[0].ID,
		PeakPrice: 10000,
		UpdatedAt: f.now,
	}
	if err := f.repo.Upsert(f.ctx, peak); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := f.repo.GetByHoldingID(f.ctx, f.holdings[0].ID)
	if err != nil {
		t.Fatalf("GetByHoldingID() error = %v", err)
	}
	if got.PeakPrice != 10000 {
		t.Errorf("PeakPrice = %v, want 10000", got.PeakPrice)
	}
	if got.HoldingID != f.holdings[0].ID {
		t.Errorf("HoldingID = %q, want %q", got.HoldingID, f.holdings[0].ID)
	}
}

func TestPeakRepoUpsertUpdatesExisting(t *testing.T) {
	f := setupPeakRepoTest(t)

	peak := &trailingstop.HoldingPeak{
		ID:        shared.NewID(),
		HoldingID: f.holdings[0].ID,
		PeakPrice: 10000,
		UpdatedAt: f.now,
	}
	if err := f.repo.Upsert(f.ctx, peak); err != nil {
		t.Fatalf("Upsert() insert error = %v", err)
	}

	peak.PeakPrice = 12000
	peak.UpdatedAt = f.now.Add(time.Hour)
	if err := f.repo.Upsert(f.ctx, peak); err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}

	got, err := f.repo.GetByHoldingID(f.ctx, f.holdings[0].ID)
	if err != nil {
		t.Fatalf("GetByHoldingID() error = %v", err)
	}
	if got.PeakPrice != 12000 {
		t.Errorf("PeakPrice = %v, want 12000", got.PeakPrice)
	}
}

func TestPeakRepoGetByHoldingIDNotFound(t *testing.T) {
	f := setupPeakRepoTest(t)

	_, err := f.repo.GetByHoldingID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByHoldingID() error = %v, want ErrNotFound", err)
	}
}

func TestPeakRepoListByHoldingIDs(t *testing.T) {
	f := setupPeakRepoTest(t)

	for i, h := range f.holdings {
		peak := &trailingstop.HoldingPeak{
			ID:        shared.NewID(),
			HoldingID: h.ID,
			PeakPrice: float64((i + 1) * 10000),
			UpdatedAt: f.now,
		}
		if err := f.repo.Upsert(f.ctx, peak); err != nil {
			t.Fatalf("Upsert() error = %v", err)
		}
	}

	ids := []string{f.holdings[0].ID, f.holdings[1].ID}
	peaks, err := f.repo.ListByHoldingIDs(f.ctx, ids)
	if err != nil {
		t.Fatalf("ListByHoldingIDs() error = %v", err)
	}
	if len(peaks) != 2 {
		t.Fatalf("len(peaks) = %d, want 2", len(peaks))
	}
}

func TestPeakRepoListByHoldingIDsEmpty(t *testing.T) {
	f := setupPeakRepoTest(t)

	peaks, err := f.repo.ListByHoldingIDs(f.ctx, nil)
	if err != nil {
		t.Fatalf("ListByHoldingIDs() error = %v", err)
	}
	if peaks != nil {
		t.Errorf("expected nil, got %v", peaks)
	}
}

func TestPeakRepoListByHoldingIDsPartialMatch(t *testing.T) {
	f := setupPeakRepoTest(t)

	peak := &trailingstop.HoldingPeak{
		ID:        shared.NewID(),
		HoldingID: f.holdings[0].ID,
		PeakPrice: 10000,
		UpdatedAt: f.now,
	}
	if err := f.repo.Upsert(f.ctx, peak); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	ids := []string{f.holdings[0].ID, f.holdings[1].ID}
	peaks, err := f.repo.ListByHoldingIDs(f.ctx, ids)
	if err != nil {
		t.Fatalf("ListByHoldingIDs() error = %v", err)
	}
	if len(peaks) != 1 {
		t.Fatalf("len(peaks) = %d, want 1", len(peaks))
	}
	if peaks[0].HoldingID != f.holdings[0].ID {
		t.Errorf("HoldingID = %q, want %q", peaks[0].HoldingID, f.holdings[0].ID)
	}
}

func TestPeakRepoTimestampRoundTrip(t *testing.T) {
	f := setupPeakRepoTest(t)

	peak := &trailingstop.HoldingPeak{
		ID:        shared.NewID(),
		HoldingID: f.holdings[0].ID,
		PeakPrice: 10000,
		UpdatedAt: f.now.Add(3 * time.Hour),
	}
	if err := f.repo.Upsert(f.ctx, peak); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := f.repo.GetByHoldingID(f.ctx, f.holdings[0].ID)
	if err != nil {
		t.Fatalf("GetByHoldingID() error = %v", err)
	}
	if !got.UpdatedAt.Equal(f.now.Add(3 * time.Hour)) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, f.now.Add(3*time.Hour))
	}
}
