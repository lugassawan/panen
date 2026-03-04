package database

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
	"github.com/lugassawan/panen/backend/domain/watchlist"
)

type tickerCollectorFixture struct {
	collector   *TickerCollector
	watchRepo   *WatchlistRepo
	itemRepo    *WatchlistItemRepo
	holdingRepo *HoldingRepo
	profileID   string
	watchlistID string
	portfolioID string
	ctx         context.Context
	now         time.Time
}

func setupTickerCollectorTest(t *testing.T) tickerCollectorFixture {
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

	wl := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: p.ID,
		Name:      "Test Watchlist",
		CreatedAt: now,
		UpdatedAt: now,
	}
	watchRepo := NewWatchlistRepo(db)
	if err := watchRepo.Create(ctx, wl); err != nil {
		t.Fatalf("create watchlist: %v", err)
	}

	return tickerCollectorFixture{
		collector:   NewTickerCollector(db),
		watchRepo:   watchRepo,
		itemRepo:    NewWatchlistItemRepo(db),
		holdingRepo: NewHoldingRepo(db),
		profileID:   p.ID,
		watchlistID: wl.ID,
		portfolioID: port.ID,
		ctx:         ctx,
		now:         now,
	}
}

func (f tickerCollectorFixture) addWatchlistItem(t *testing.T, ticker string) {
	t.Helper()
	item := &watchlist.Item{
		ID:          shared.NewID(),
		WatchlistID: f.watchlistID,
		Ticker:      ticker,
		CreatedAt:   f.now,
	}
	if err := f.itemRepo.Add(f.ctx, item); err != nil {
		t.Fatalf("add watchlist item %s: %v", ticker, err)
	}
}

func (f tickerCollectorFixture) addHolding(t *testing.T, ticker string) {
	t.Helper()
	h := &portfolio.Holding{
		ID:          shared.NewID(),
		PortfolioID: f.portfolioID,
		Ticker:      ticker,
		AvgBuyPrice: 1000,
		Lots:        1,
		CreatedAt:   f.now,
		UpdatedAt:   f.now,
	}
	if err := f.holdingRepo.Create(f.ctx, h); err != nil {
		t.Fatalf("add holding %s: %v", ticker, err)
	}
}

func TestTickerCollectorEmpty(t *testing.T) {
	f := setupTickerCollectorTest(t)

	// Remove the watchlist so we have truly empty items
	// (the fixture creates a watchlist but no items or holdings)
	got, err := f.collector.CollectAll(f.ctx)
	if err != nil {
		t.Fatalf("CollectAll() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("CollectAll() = %v, want empty slice", got)
	}
}

func TestTickerCollectorWatchlistOnly(t *testing.T) {
	f := setupTickerCollectorTest(t)

	f.addWatchlistItem(t, "BBCA")
	f.addWatchlistItem(t, "TLKM")

	got, err := f.collector.CollectAll(f.ctx)
	if err != nil {
		t.Fatalf("CollectAll() error = %v", err)
	}
	want := []string{"BBCA", "TLKM"}
	if len(got) != len(want) {
		t.Fatalf("CollectAll() count = %d, want %d", len(got), len(want))
	}
	for i, ticker := range want {
		if got[i] != ticker {
			t.Errorf("CollectAll()[%d] = %q, want %q", i, got[i], ticker)
		}
	}
}

func TestTickerCollectorHoldingsOnly(t *testing.T) {
	f := setupTickerCollectorTest(t)

	f.addHolding(t, "BMRI")
	f.addHolding(t, "ASII")

	got, err := f.collector.CollectAll(f.ctx)
	if err != nil {
		t.Fatalf("CollectAll() error = %v", err)
	}
	want := []string{"ASII", "BMRI"}
	if len(got) != len(want) {
		t.Fatalf("CollectAll() count = %d, want %d", len(got), len(want))
	}
	for i, ticker := range want {
		if got[i] != ticker {
			t.Errorf("CollectAll()[%d] = %q, want %q", i, got[i], ticker)
		}
	}
}

func TestTickerCollectorDeduplication(t *testing.T) {
	f := setupTickerCollectorTest(t)

	// BBCA in both watchlist and holdings
	f.addWatchlistItem(t, "BBCA")
	f.addWatchlistItem(t, "TLKM")
	f.addHolding(t, "BBCA")
	f.addHolding(t, "BMRI")

	got, err := f.collector.CollectAll(f.ctx)
	if err != nil {
		t.Fatalf("CollectAll() error = %v", err)
	}
	want := []string{"BBCA", "BMRI", "TLKM"}
	if len(got) != len(want) {
		t.Fatalf("CollectAll() count = %d, want %d; got %v", len(got), len(want), got)
	}
	for i, ticker := range want {
		if got[i] != ticker {
			t.Errorf("CollectAll()[%d] = %q, want %q", i, got[i], ticker)
		}
	}
}
