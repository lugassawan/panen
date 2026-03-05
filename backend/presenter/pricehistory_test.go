package presenter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/usecase"
)

type mockPriceHistoryRepo struct {
	mu     sync.Mutex
	points map[string][]stock.PricePoint
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
	return time.Time{}, nil
}

func (r *mockPriceHistoryRepo) DeleteByTicker(_ context.Context, _ string) error {
	return nil
}

func newTestPriceHistoryHandler(
	historyFunc func(ctx context.Context, ticker string) ([]stock.PricePoint, error),
) *PriceHistoryHandler {
	repo := newMockPriceHistoryRepo()
	provider := &mockPriceHistoryProvider{
		source:      "mock",
		historyFunc: historyFunc,
	}
	svc := usecase.NewPriceHistoryService(repo, provider)
	return NewPriceHistoryHandler(context.Background(), svc)
}

type mockPriceHistoryProvider struct {
	source      string
	historyFunc func(ctx context.Context, ticker string) ([]stock.PricePoint, error)
}

func (m *mockPriceHistoryProvider) Source() string { return m.source }

func (m *mockPriceHistoryProvider) FetchPrice(_ context.Context, _ string) (*stock.PriceResult, error) {
	return &stock.PriceResult{}, nil
}

func (m *mockPriceHistoryProvider) FetchFinancials(_ context.Context, _ string) (*stock.FinancialResult, error) {
	return &stock.FinancialResult{}, nil
}

func (m *mockPriceHistoryProvider) FetchPriceHistory(ctx context.Context, ticker string) ([]stock.PricePoint, error) {
	if m.historyFunc != nil {
		return m.historyFunc(ctx, ticker)
	}
	return nil, nil
}

func TestGetPriceHistory(t *testing.T) {
	date := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	handler := newTestPriceHistoryHandler(func(_ context.Context, ticker string) ([]stock.PricePoint, error) {
		return []stock.PricePoint{
			{
				Ticker: ticker,
				Date:   date,
				Open:   9000,
				High:   9200,
				Low:    8900,
				Close:  9100,
				Volume: 100000,
				Source: "mock",
			},
			{
				Ticker: ticker,
				Date:   date.AddDate(0, 0, 1),
				Open:   9100,
				High:   9300,
				Low:    9000,
				Close:  9250,
				Volume: 150000,
				Source: "mock",
			},
		}, nil
	})

	result, err := handler.GetPriceHistory("BBCA")
	if err != nil {
		t.Fatalf("GetPriceHistory() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].Date != "2025-03-01" {
		t.Errorf("result[0].Date = %q, want 2025-03-01", result[0].Date)
	}
	if result[0].Close != 9100 {
		t.Errorf("result[0].Close = %v, want 9100", result[0].Close)
	}
	if result[1].Volume != 150000 {
		t.Errorf("result[1].Volume = %v, want 150000", result[1].Volume)
	}
}

func TestGetPriceHistoryError(t *testing.T) {
	handler := newTestPriceHistoryHandler(func(_ context.Context, _ string) ([]stock.PricePoint, error) {
		return nil, errors.New("provider error")
	})

	_, err := handler.GetPriceHistory("BBCA")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
