package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
)

// mockPriceHistoryRepo is an in-memory stock.PriceHistoryRepository for testing.
type mockPriceHistoryRepo struct {
	mu         sync.Mutex
	points     map[string][]stock.PricePoint // keyed by "ticker:source"
	latestDate time.Time
}

func newMockPriceHistoryRepo() *mockPriceHistoryRepo {
	return &mockPriceHistoryRepo{points: make(map[string][]stock.PricePoint)}
}

func (r *mockPriceHistoryRepo) BulkUpsert(_ context.Context, points []stock.PricePoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range points {
		key := p.Ticker + ":" + p.Source
		r.points[key] = append(r.points[key], p)
	}
	return nil
}

func (r *mockPriceHistoryRepo) GetByTicker(_ context.Context, ticker, source string) ([]stock.PricePoint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.points[ticker+":"+source], nil
}

func (r *mockPriceHistoryRepo) LatestDate(_ context.Context, _, _ string) (time.Time, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.latestDate, nil
}

func (r *mockPriceHistoryRepo) DeleteByTicker(_ context.Context, ticker string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key := range r.points {
		if len(key) > len(ticker) && key[:len(ticker)] == ticker {
			delete(r.points, key)
		}
	}
	return nil
}

func TestPriceHistoryServiceGetHistoryFetchesWhenEmpty(t *testing.T) {
	repo := newMockPriceHistoryRepo()
	provider := newMockProvider()
	fetchCalled := false
	provider.source = "mock"

	svc := &PriceHistoryService{
		historyRepo: repo,
		provider: &mockPriceHistoryProvider{
			mockProvider: provider,
			historyFunc: func(_ context.Context, ticker string) ([]stock.PricePoint, error) {
				fetchCalled = true
				return []stock.PricePoint{
					{Ticker: ticker, Date: time.Now().UTC(), Close: 9100, Source: "mock"},
				}, nil
			},
		},
	}

	points, err := svc.GetHistory(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if !fetchCalled {
		t.Error("expected provider.FetchPriceHistory to be called")
	}
	if len(points) != 1 {
		t.Errorf("len(points) = %d, want 1", len(points))
	}
}

func TestPriceHistoryServiceGetHistoryUsesCacheWhenFresh(t *testing.T) {
	repo := newMockPriceHistoryRepo()
	repo.latestDate = time.Now().UTC()
	repo.points["BBCA:mock"] = []stock.PricePoint{
		{Ticker: "BBCA", Date: time.Now().UTC(), Close: 9100, Source: "mock"},
	}

	provider := newMockProvider()
	fetchCalled := false

	svc := &PriceHistoryService{
		historyRepo: repo,
		provider: &mockPriceHistoryProvider{
			mockProvider: provider,
			historyFunc: func(_ context.Context, _ string) ([]stock.PricePoint, error) {
				fetchCalled = true
				return nil, errors.New("should not be called")
			},
		},
	}

	points, err := svc.GetHistory(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if fetchCalled {
		t.Error("expected provider.FetchPriceHistory NOT to be called for fresh data")
	}
	if len(points) != 1 {
		t.Errorf("len(points) = %d, want 1", len(points))
	}
}

func TestPriceHistoryServiceGetHistoryRefetchesWhenStale(t *testing.T) {
	repo := newMockPriceHistoryRepo()
	repo.latestDate = time.Now().UTC().AddDate(0, 0, -3) // 3 days old

	provider := newMockProvider()
	fetchCalled := false

	svc := &PriceHistoryService{
		historyRepo: repo,
		provider: &mockPriceHistoryProvider{
			mockProvider: provider,
			historyFunc: func(_ context.Context, ticker string) ([]stock.PricePoint, error) {
				fetchCalled = true
				return []stock.PricePoint{
					{Ticker: ticker, Date: time.Now().UTC(), Close: 9500, Source: "mock"},
				}, nil
			},
		},
	}

	_, err := svc.GetHistory(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if !fetchCalled {
		t.Error("expected provider.FetchPriceHistory to be called for stale data")
	}
}

func TestPriceHistoryServiceGetHistoryPropagatesProviderError(t *testing.T) {
	repo := newMockPriceHistoryRepo()

	svc := &PriceHistoryService{
		historyRepo: repo,
		provider: &mockPriceHistoryProvider{
			mockProvider: newMockProvider(),
			historyFunc: func(_ context.Context, _ string) ([]stock.PricePoint, error) {
				return nil, stock.ErrNoData
			},
		},
	}

	_, err := svc.GetHistory(context.Background(), "BBCA")
	if !errors.Is(err, stock.ErrNoData) {
		t.Errorf("expected ErrNoData, got: %v", err)
	}
}

// mockPriceHistoryProvider wraps mockProvider and adds FetchPriceHistory.
type mockPriceHistoryProvider struct {
	*mockProvider
	historyFunc func(ctx context.Context, ticker string) ([]stock.PricePoint, error)
}

func (m *mockPriceHistoryProvider) FetchPriceHistory(ctx context.Context, ticker string) ([]stock.PricePoint, error) {
	if m.historyFunc != nil {
		return m.historyFunc(ctx, ticker)
	}
	return nil, nil
}
