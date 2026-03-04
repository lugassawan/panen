package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

// EventEmitter abstracts event emission (implemented by Wails runtime wrapper in presenter).
type EventEmitter interface {
	Emit(eventName string, data any)
}

// TickerCollector collects tracked tickers from watchlists and holdings.
type TickerCollector interface {
	CollectAll(ctx context.Context) ([]string, error)
}

// RefreshProgress reports progress for a single ticker during a refresh cycle.
type RefreshProgress struct {
	Ticker string `json:"ticker"`
	Index  int    `json:"index"`
	Total  int    `json:"total"`
	Status string `json:"status"` // "success", "skipped", "error"
	Error  string `json:"error,omitempty"`
}

// RefreshSummary reports the outcome of a completed refresh cycle.
type RefreshSummary struct {
	Total    int    `json:"total"`
	Fetched  int    `json:"fetched"`
	Skipped  int    `json:"skipped"`
	Failed   int    `json:"failed"`
	Duration string `json:"duration"` // human-readable, e.g. "2.3s"
}

// RefreshStatus represents the current state of the refresh service.
type RefreshStatus struct {
	State       string `json:"state"`       // "idle", "syncing", "error"
	LastRefresh string `json:"lastRefresh"` // ISO 8601 or empty
	Error       string `json:"error,omitempty"`
}

const (
	maxRetries      = 3
	retryBaseWait   = 1 * time.Second
	defaultInterval = 720 * time.Minute
)

// RefreshService manages background auto-refresh of stock data.
type RefreshService struct {
	stockData stock.Repository
	provider  stock.DataProvider
	settings  settings.Repository
	collector TickerCollector
	emitter   EventEmitter

	mu      sync.Mutex
	status  RefreshStatus
	running atomic.Bool
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewRefreshService creates a new RefreshService.
func NewRefreshService(
	stockData stock.Repository,
	provider stock.DataProvider,
	settings settings.Repository,
	collector TickerCollector,
	emitter EventEmitter,
) *RefreshService {
	return &RefreshService{
		stockData: stockData,
		provider:  provider,
		settings:  settings,
		collector: collector,
		emitter:   emitter,
		status:    RefreshStatus{State: "idle"},
	}
}

// Start launches a background goroutine that refreshes stock data on a timer.
// It runs one immediate refresh on startup, then refreshes at the configured interval.
func (r *RefreshService) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	r.done = make(chan struct{})

	go r.loop(ctx)
}

// Stop cancels the background goroutine and waits for it to finish.
func (r *RefreshService) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	if r.done != nil {
		<-r.done
	}
}

// RunNow triggers an immediate refresh cycle.
func (r *RefreshService) RunNow(ctx context.Context) error {
	return r.refresh(ctx)
}

// GetStatus returns the current refresh status (thread-safe).
func (r *RefreshService) GetStatus() RefreshStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status
}

func (r *RefreshService) readInterval() time.Duration {
	cfg, err := r.settings.GetRefreshSettings(context.Background())
	if err != nil {
		return defaultInterval
	}
	if d := time.Duration(cfg.IntervalMinutes) * time.Minute; d > 0 {
		return d
	}
	return defaultInterval
}

func (r *RefreshService) loop(ctx context.Context) {
	defer close(r.done)

	_ = r.refresh(ctx)

	interval := r.readInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cfg, err := r.settings.GetRefreshSettings(ctx)
			if err != nil || !cfg.AutoRefreshEnabled {
				continue
			}

			newInterval := time.Duration(cfg.IntervalMinutes) * time.Minute
			if newInterval > 0 && newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
			}

			_ = r.refresh(ctx)
		}
	}
}

func (r *RefreshService) setStatus(s RefreshStatus) {
	r.mu.Lock()
	r.status = s
	r.mu.Unlock()
	r.emitter.Emit("refresh:status", s)
}

func (r *RefreshService) refresh(ctx context.Context) error {
	if !r.running.CompareAndSwap(false, true) {
		return nil // another refresh is already in progress
	}
	defer r.running.Store(false)

	r.setStatus(RefreshStatus{State: "syncing"})

	start := time.Now()

	tickers, err := r.collector.CollectAll(ctx)
	if err != nil {
		errStatus := RefreshStatus{State: "error", Error: fmt.Sprintf("collect tickers: %v", err)}
		r.setStatus(errStatus)
		return fmt.Errorf("collect tickers: %w", err)
	}

	total := len(tickers)
	var fetched, skipped, failed int

	for i, ticker := range tickers {
		if ctx.Err() != nil {
			break
		}

		progress := r.refreshTicker(ctx, ticker)
		switch progress.Status {
		case "success":
			fetched++
		case "skipped":
			skipped++
		default:
			failed++
		}
		progress.Index = i
		progress.Total = total
		r.emitter.Emit("refresh:progress", progress)
	}

	duration := time.Since(start)
	r.emitter.Emit("refresh:summary", RefreshSummary{
		Total:    total,
		Fetched:  fetched,
		Skipped:  skipped,
		Failed:   failed,
		Duration: formatDuration(duration),
	})

	// Save last refreshed time.
	now := time.Now().UTC()
	cfg, err := r.settings.GetRefreshSettings(ctx)
	if err == nil {
		cfg.LastRefreshedAt = now
		_ = r.settings.SaveRefreshSettings(ctx, cfg)
	}

	finalState := "idle"
	if total > 0 && failed == total {
		finalState = "error"
	}
	r.setStatus(RefreshStatus{
		State:       finalState,
		LastRefresh: now.Format("2006-01-02T15:04:05Z"),
	})

	return nil
}

func (r *RefreshService) refreshTicker(ctx context.Context, ticker string) RefreshProgress {
	existing, err := r.stockData.GetByTickerAndSource(ctx, ticker, r.provider.Source())
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return RefreshProgress{Ticker: ticker, Status: "error", Error: err.Error()}
	}
	if existing != nil && time.Since(existing.FetchedAt) < cacheTTL {
		return RefreshProgress{Ticker: ticker, Status: "skipped"}
	}
	if err := r.fetchWithRetry(ctx, ticker); err != nil {
		return RefreshProgress{Ticker: ticker, Status: "error", Error: err.Error()}
	}
	return RefreshProgress{Ticker: ticker, Status: "success"}
}

func (r *RefreshService) fetchWithRetry(ctx context.Context, ticker string) error {
	var lastErr error
	for attempt := range maxRetries {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := r.fetchAndUpsert(ctx, ticker); err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				wait := retryBaseWait * (1 << uint(attempt)) // 1s, 2s, 4s
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(wait):
				}
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}

func (r *RefreshService) fetchAndUpsert(ctx context.Context, ticker string) error {
	price, err := r.provider.FetchPrice(ctx, ticker)
	if err != nil {
		return fmt.Errorf("fetch price %s: %w", ticker, err)
	}

	fin, err := r.provider.FetchFinancials(ctx, ticker)
	if err != nil {
		return fmt.Errorf("fetch financials %s: %w", ticker, err)
	}

	data := &stock.Data{
		ID:            shared.NewID(),
		Ticker:        ticker,
		Price:         price.Price,
		High52Week:    price.High52Week,
		Low52Week:     price.Low52Week,
		EPS:           fin.EPS,
		BVPS:          fin.BVPS,
		ROE:           fin.ROE,
		DER:           fin.DER,
		PBV:           fin.PBV,
		PER:           fin.PER,
		DividendYield: fin.DividendYield,
		PayoutRatio:   fin.PayoutRatio,
		FetchedAt:     time.Now().UTC(),
		Source:        r.provider.Source(),
	}

	return r.stockData.Upsert(ctx, data)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
