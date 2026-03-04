package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/stock"
)

// helper to build a RefreshService with common defaults.
func newTestRefreshService(
	repo *mockStockRepo,
	provider *mockProvider,
	settingsRepo *mockSettingsRepo,
	collector *mockTickerCollector,
	emitter *mockEventEmitter,
) *RefreshService {
	return NewRefreshService(repo, provider, settingsRepo, collector, emitter)
}

func TestRefreshService(t *testing.T) {
	type setupFn func(
		repo *mockStockRepo, provider *mockProvider,
		settingsRepo *mockSettingsRepo, collector *mockTickerCollector,
	)

	tests := []struct {
		name        string
		setup       setupFn
		wantFetched int
		wantSkipped int
		wantFailed  int
		wantState   string
		wantCalls   int // provider FetchPrice call count
	}{
		{
			name: "fresh data skipped",
			setup: func(repo *mockStockRepo, provider *mockProvider, _ *mockSettingsRepo, collector *mockTickerCollector) {
				collector.tickers = []string{"BBCA", "BBRI"}
				// Seed fresh data for both tickers.
				for _, ticker := range collector.tickers {
					_ = repo.Upsert(context.Background(), &stock.Data{
						ID: ticker + "-id", Ticker: ticker, Price: 8000,
						FetchedAt: time.Now().UTC(), Source: provider.Source(),
					})
				}
			},
			wantFetched: 0,
			wantSkipped: 2,
			wantFailed:  0,
			wantState:   "idle",
			wantCalls:   0,
		},
		{
			name: "stale data refreshed",
			setup: func(repo *mockStockRepo, provider *mockProvider, _ *mockSettingsRepo, collector *mockTickerCollector) {
				collector.tickers = []string{"BBCA"}
				// Seed stale data (older than 24h).
				_ = repo.Upsert(context.Background(), &stock.Data{
					ID: "stale-id", Ticker: "BBCA", Price: 7500,
					FetchedAt: time.Now().UTC().Add(-25 * time.Hour), Source: provider.Source(),
				})
			},
			wantFetched: 1,
			wantSkipped: 0,
			wantFailed:  0,
			wantState:   "idle",
			wantCalls:   1,
		},
		{
			name: "mixed fresh and stale",
			setup: func(repo *mockStockRepo, provider *mockProvider, _ *mockSettingsRepo, collector *mockTickerCollector) {
				collector.tickers = []string{"BBCA", "BBRI", "TLKM"}
				// BBCA: fresh
				_ = repo.Upsert(context.Background(), &stock.Data{
					ID: "fresh-id", Ticker: "BBCA", Price: 8000,
					FetchedAt: time.Now().UTC(), Source: provider.Source(),
				})
				// BBRI: stale
				_ = repo.Upsert(context.Background(), &stock.Data{
					ID: "stale-id", Ticker: "BBRI", Price: 4500,
					FetchedAt: time.Now().UTC().Add(-25 * time.Hour), Source: provider.Source(),
				})
				// TLKM: not in repo at all (new)
			},
			wantFetched: 2,
			wantSkipped: 1,
			wantFailed:  0,
			wantState:   "idle",
			wantCalls:   2,
		},
		{
			name: "empty ticker list",
			setup: func(_ *mockStockRepo, _ *mockProvider, _ *mockSettingsRepo, collector *mockTickerCollector) {
				collector.tickers = nil
			},
			wantFetched: 0,
			wantSkipped: 0,
			wantFailed:  0,
			wantState:   "idle",
			wantCalls:   0,
		},
		{
			name: "permanent failure after retries",
			setup: func(_ *mockStockRepo, provider *mockProvider, _ *mockSettingsRepo, collector *mockTickerCollector) {
				collector.tickers = []string{"FAIL"}
				provider.priceFunc = func(_ context.Context, _ string) (*stock.PriceResult, error) {
					return nil, errors.New("network error")
				}
			},
			wantFetched: 0,
			wantSkipped: 0,
			wantFailed:  1,
			wantState:   "error", // all tickers failed
			wantCalls:   3,       // 3 retry attempts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockStockRepo()
			provider := newMockProvider()
			settingsRepo := newMockSettingsRepo()
			collector := newMockTickerCollector()
			emitter := newMockEventEmitter()

			tt.setup(repo, provider, settingsRepo, collector)

			svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

			err := svc.RunNow(context.Background())
			if err != nil {
				t.Fatalf("RunNow() error = %v", err)
			}

			// Check summary event.
			summaries := emitter.eventsByName("refresh:summary")
			if len(summaries) != 1 {
				t.Fatalf("expected 1 refresh:summary event, got %d", len(summaries))
			}
			summary, ok := summaries[0].data.(RefreshSummary)
			if !ok {
				t.Fatal("refresh:summary data is not RefreshSummary")
			}
			if summary.Fetched != tt.wantFetched {
				t.Errorf("Fetched = %d, want %d", summary.Fetched, tt.wantFetched)
			}
			if summary.Skipped != tt.wantSkipped {
				t.Errorf("Skipped = %d, want %d", summary.Skipped, tt.wantSkipped)
			}
			if summary.Failed != tt.wantFailed {
				t.Errorf("Failed = %d, want %d", summary.Failed, tt.wantFailed)
			}

			// Check final status.
			status := svc.GetStatus()
			if status.State != tt.wantState {
				t.Errorf("State = %q, want %q", status.State, tt.wantState)
			}

			// Check provider call count.
			provider.mu.Lock()
			calls := provider.callCount
			provider.mu.Unlock()
			if calls != tt.wantCalls {
				t.Errorf("FetchPrice called %d times, want %d", calls, tt.wantCalls)
			}
		})
	}
}

func TestRefreshServiceRetryThenSucceed(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	settingsRepo := newMockSettingsRepo()
	collector := newMockTickerCollector("BBCA")
	emitter := newMockEventEmitter()

	// Fail on first attempt, succeed on second.
	var attempts int
	provider.priceFunc = func(_ context.Context, _ string) (*stock.PriceResult, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("temporary error")
		}
		return &stock.PriceResult{Price: 8500, High52Week: 9000, Low52Week: 7000}, nil
	}

	svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

	err := svc.RunNow(context.Background())
	if err != nil {
		t.Fatalf("RunNow() error = %v", err)
	}

	summaries := emitter.eventsByName("refresh:summary")
	if len(summaries) != 1 {
		t.Fatalf("expected 1 refresh:summary event, got %d", len(summaries))
	}
	summary, ok := summaries[0].data.(RefreshSummary)
	if !ok {
		t.Fatal("refresh:summary data is not RefreshSummary")
	}
	if summary.Fetched != 1 {
		t.Errorf("Fetched = %d, want 1", summary.Fetched)
	}
	if summary.Failed != 0 {
		t.Errorf("Failed = %d, want 0", summary.Failed)
	}

	// Should have been called twice: first failure + second success.
	provider.mu.Lock()
	calls := provider.callCount
	provider.mu.Unlock()
	if calls != 2 {
		t.Errorf("FetchPrice called %d times, want 2 (1 fail + 1 success)", calls)
	}

	// Verify data was upserted.
	data, dataErr := repo.GetByTickerAndSource(context.Background(), "BBCA", "mock")
	if dataErr != nil {
		t.Fatalf("GetByTickerAndSource() error = %v", dataErr)
	}
	if data.Price != 8500 {
		t.Errorf("Price = %f, want 8500", data.Price)
	}
}

func TestRefreshServiceContextCancellation(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	settingsRepo := newMockSettingsRepo()
	collector := newMockTickerCollector("BBCA", "BBRI", "TLKM")
	emitter := newMockEventEmitter()

	// Cancel context after first fetch.
	ctx, cancel := context.WithCancel(context.Background())
	var fetchCount int
	provider.priceFunc = func(_ context.Context, _ string) (*stock.PriceResult, error) {
		fetchCount++
		if fetchCount >= 1 {
			cancel()
		}
		return &stock.PriceResult{Price: 8500, High52Week: 9000, Low52Week: 7000}, nil
	}

	svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

	// RunNow should not return an error; it handles cancellation gracefully.
	_ = svc.RunNow(ctx)

	// Not all tickers should have been fetched due to cancellation.
	provider.mu.Lock()
	calls := provider.callCount
	provider.mu.Unlock()

	if calls >= 3 {
		t.Errorf("FetchPrice called %d times, expected fewer due to cancellation", calls)
	}
}

func TestRefreshServiceAutoRefreshDisabled(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	settingsRepo := newMockSettingsRepo()
	collector := newMockTickerCollector("BBCA")
	emitter := newMockEventEmitter()

	// Disable auto-refresh.
	settingsRepo.mu.Lock()
	settingsRepo.settings.AutoRefreshEnabled = false
	settingsRepo.mu.Unlock()

	svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

	// RunNow still works regardless of auto-refresh setting (it's a manual trigger).
	// The auto-refresh disabled check only applies to the background loop.
	// We test the loop behavior by verifying Start/Stop with disabled setting.
	ctx, cancel := context.WithCancel(context.Background())
	svc.Start(ctx)

	// Give the loop time to run the initial refresh and one tick.
	time.Sleep(50 * time.Millisecond)
	cancel()
	svc.Stop()

	// The initial refresh in Start() always runs (1 call).
	// But subsequent ticks should skip because auto-refresh is disabled.
	provider.mu.Lock()
	calls := provider.callCount
	provider.mu.Unlock()

	// Should have exactly 1 call from the initial startup refresh.
	if calls != 1 {
		t.Errorf("FetchPrice called %d times, want 1 (only startup refresh)", calls)
	}
}

func TestRefreshServiceStatusEvents(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	settingsRepo := newMockSettingsRepo()
	collector := newMockTickerCollector("BBCA")
	emitter := newMockEventEmitter()

	svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

	err := svc.RunNow(context.Background())
	if err != nil {
		t.Fatalf("RunNow() error = %v", err)
	}

	// Should have emitted at least 2 status events: syncing and idle.
	statusEvents := emitter.eventsByName("refresh:status")
	if len(statusEvents) < 2 {
		t.Fatalf("expected at least 2 refresh:status events, got %d", len(statusEvents))
	}

	first, ok := statusEvents[0].data.(RefreshStatus)
	if !ok {
		t.Fatal("first refresh:status data is not RefreshStatus")
	}
	if first.State != "syncing" {
		t.Errorf("first status State = %q, want %q", first.State, "syncing")
	}

	last, ok := statusEvents[len(statusEvents)-1].data.(RefreshStatus)
	if !ok {
		t.Fatal("last refresh:status data is not RefreshStatus")
	}
	if last.State != "idle" {
		t.Errorf("last status State = %q, want %q", last.State, "idle")
	}
	if last.LastRefresh == "" {
		t.Error("last status LastRefresh should not be empty")
	}
}

func TestRefreshServiceProgressEvents(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	settingsRepo := newMockSettingsRepo()
	collector := newMockTickerCollector("BBCA", "BBRI")
	emitter := newMockEventEmitter()

	// BBCA: fresh (will be skipped), BBRI: not in repo (will be fetched)
	_ = repo.Upsert(context.Background(), &stock.Data{
		ID: "fresh-id", Ticker: "BBCA", Price: 8000,
		FetchedAt: time.Now().UTC(), Source: provider.Source(),
	})

	svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

	err := svc.RunNow(context.Background())
	if err != nil {
		t.Fatalf("RunNow() error = %v", err)
	}

	progressEvents := emitter.eventsByName("refresh:progress")
	if len(progressEvents) != 2 {
		t.Fatalf("expected 2 refresh:progress events, got %d", len(progressEvents))
	}

	// First event: BBCA skipped.
	p0, ok := progressEvents[0].data.(RefreshProgress)
	if !ok {
		t.Fatal("progress event data is not RefreshProgress")
	}
	if p0.Ticker != "BBCA" {
		t.Errorf("progress[0] Ticker = %q, want BBCA", p0.Ticker)
	}
	if p0.Status != "skipped" {
		t.Errorf("progress[0] Status = %q, want skipped", p0.Status)
	}
	if p0.Total != 2 {
		t.Errorf("progress[0] Total = %d, want 2", p0.Total)
	}

	// Second event: BBRI success.
	p1, ok := progressEvents[1].data.(RefreshProgress)
	if !ok {
		t.Fatal("progress event data is not RefreshProgress")
	}
	if p1.Ticker != "BBRI" {
		t.Errorf("progress[1] Ticker = %q, want BBRI", p1.Ticker)
	}
	if p1.Status != "success" {
		t.Errorf("progress[1] Status = %q, want success", p1.Status)
	}
}

func TestRefreshServiceSavesLastRefreshedAt(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	settingsRepo := newMockSettingsRepo()
	collector := newMockTickerCollector("BBCA")
	emitter := newMockEventEmitter()

	before := time.Now().UTC()

	svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

	err := svc.RunNow(context.Background())
	if err != nil {
		t.Fatalf("RunNow() error = %v", err)
	}

	settingsRepo.mu.Lock()
	lastRefresh := settingsRepo.settings.LastRefreshedAt
	settingsRepo.mu.Unlock()

	if lastRefresh.Before(before) {
		t.Errorf("LastRefreshedAt = %v, expected after %v", lastRefresh, before)
	}
}

func TestRefreshServiceStartStop(t *testing.T) {
	repo := newMockStockRepo()
	provider := newMockProvider()
	settingsRepo := newMockSettingsRepo()
	collector := newMockTickerCollector("BBCA")
	emitter := newMockEventEmitter()

	svc := newTestRefreshService(repo, provider, settingsRepo, collector, emitter)

	ctx := context.Background()
	svc.Start(ctx)

	// Give the initial refresh time to complete.
	time.Sleep(50 * time.Millisecond)

	svc.Stop()

	// Verify the initial refresh ran.
	provider.mu.Lock()
	calls := provider.callCount
	provider.mu.Unlock()
	if calls < 1 {
		t.Errorf("FetchPrice called %d times, want at least 1 (initial refresh)", calls)
	}

	// Verify status is idle after stop.
	status := svc.GetStatus()
	if status.State != "idle" {
		t.Errorf("State = %q after Stop(), want idle", status.State)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{2300 * time.Millisecond, "2.3s"},
		{100 * time.Millisecond, "100ms"},
		{1 * time.Second, "1.0s"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
