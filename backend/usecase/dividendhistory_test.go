package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

type mockDividendHistoryRepo struct {
	mu     sync.Mutex
	events map[string][]dividend.DividendEvent
	latest time.Time
}

func newMockDividendHistoryRepo() *mockDividendHistoryRepo {
	return &mockDividendHistoryRepo{events: make(map[string][]dividend.DividendEvent)}
}

func (r *mockDividendHistoryRepo) BulkUpsert(_ context.Context, events []dividend.DividendEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range events {
		key := e.Ticker + ":" + e.Source
		r.events[key] = append(r.events[key], e)
	}
	return nil
}

func (r *mockDividendHistoryRepo) GetByTicker(
	_ context.Context,
	ticker, source string,
) ([]dividend.DividendEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.events[ticker+":"+source], nil
}

func (r *mockDividendHistoryRepo) LatestDate(_ context.Context, _, _ string) (time.Time, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.latest, nil
}

type mockDividendProvider struct {
	*mockProvider
	divHistoryFunc func(ctx context.Context, ticker string) ([]dividend.DividendEvent, error)
}

func (m *mockDividendProvider) FetchDividendHistory(
	ctx context.Context,
	ticker string,
) ([]dividend.DividendEvent, error) {
	if m.divHistoryFunc != nil {
		return m.divHistoryFunc(ctx, ticker)
	}
	return nil, nil
}

func TestDividendHistoryServiceGetDividendHistory(t *testing.T) {
	repo := newMockDividendHistoryRepo()
	provider := &mockDividendProvider{
		mockProvider: newMockProvider(),
		divHistoryFunc: func(_ context.Context, ticker string) ([]dividend.DividendEvent, error) {
			return []dividend.DividendEvent{
				{Ticker: ticker, ExDate: time.Now().UTC(), Amount: 50, Source: "mock"},
			}, nil
		},
	}

	svc := NewDividendHistoryService(
		repo, provider,
		newMockHoldingRepo(),
		newMockPortfolioRepo(),
		newMockStockRepo(),
	)

	events, err := svc.GetDividendHistory(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("GetDividendHistory() error = %v", err)
	}
	if len(events) != 1 {
		t.Errorf("len(events) = %d, want 1", len(events))
	}
}

func TestDividendHistoryServiceGetDGR(t *testing.T) {
	repo := newMockDividendHistoryRepo()
	now := time.Now().UTC()
	repo.latest = now
	repo.events["BBCA:mock"] = []dividend.DividendEvent{
		{Ticker: "BBCA", ExDate: time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC), Amount: 50, Source: "mock"},
		{Ticker: "BBCA", ExDate: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC), Amount: 60, Source: "mock"},
		{Ticker: "BBCA", ExDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), Amount: 72, Source: "mock"},
	}

	provider := &mockDividendProvider{mockProvider: newMockProvider()}
	svc := NewDividendHistoryService(repo, provider, newMockHoldingRepo(), newMockPortfolioRepo(), newMockStockRepo())

	results, err := svc.GetDGR(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("GetDGR() error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	if results[1].GrowthPct != 20 {
		t.Errorf("2023 growth = %v, want 20", results[1].GrowthPct)
	}
}

func TestDividendHistoryServiceGetIncomeSummary(t *testing.T) {
	now := time.Now().UTC()
	repo := newMockDividendHistoryRepo()
	repo.latest = now
	repo.events["BBCA:mock"] = []dividend.DividendEvent{
		{Ticker: "BBCA", ExDate: now.AddDate(0, -3, 0), Amount: 50, Source: "mock"},
	}

	holdingRepo := newMockHoldingRepo()
	_ = holdingRepo.Create(context.Background(), &portfolio.Holding{
		ID:          shared.NewID(),
		PortfolioID: "p1",
		Ticker:      "BBCA",
		AvgBuyPrice: 8000,
		Lots:        10,
	})

	stockRepo := newMockStockRepo()
	_ = stockRepo.Upsert(context.Background(), &stock.Data{
		ID:            shared.NewID(),
		Ticker:        "BBCA",
		DividendYield: 3.0,
		Source:        "mock",
	})

	provider := &mockDividendProvider{mockProvider: newMockProvider()}
	svc := NewDividendHistoryService(repo, provider, holdingRepo, newMockPortfolioRepo(), stockRepo)

	summary, err := svc.GetDividendIncomeSummary(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetDividendIncomeSummary() error = %v", err)
	}
	if summary.TotalAnnualIncome <= 0 {
		t.Error("expected positive total annual income")
	}
	if len(summary.PerStock) != 1 {
		t.Errorf("len(PerStock) = %d, want 1", len(summary.PerStock))
	}
}

func TestDividendHistoryServiceGetCalendar(t *testing.T) {
	repo := newMockDividendHistoryRepo()
	repo.latest = time.Now().UTC()
	repo.events["BBCA:mock"] = []dividend.DividendEvent{
		{Ticker: "BBCA", ExDate: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50, Source: "mock"},
		{Ticker: "BBCA", ExDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 55, Source: "mock"},
	}

	holdingRepo := newMockHoldingRepo()
	_ = holdingRepo.Create(context.Background(), &portfolio.Holding{
		ID:          shared.NewID(),
		PortfolioID: "p1",
		Ticker:      "BBCA",
		Lots:        10,
	})

	provider := &mockDividendProvider{mockProvider: newMockProvider()}
	svc := NewDividendHistoryService(repo, provider, holdingRepo, newMockPortfolioRepo(), newMockStockRepo())

	projections, err := svc.GetDividendCalendar(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetDividendCalendar() error = %v", err)
	}
	if len(projections) == 0 {
		t.Error("expected projections, got 0")
	}
}
