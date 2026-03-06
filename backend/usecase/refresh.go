package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lugassawan/panen/backend/domain/alert"
	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

const (
	eventRefreshError   = "refresh:error"
	eventRefreshSummary = "refresh:summary"

	stateIdle    = "idle"
	stateSyncing = "syncing"
	stateError   = "error"
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

const snapshotKeepN = 10

// RefreshService manages background auto-refresh of stock data.
type RefreshService struct {
	stockData stock.Repository
	provider  stock.DataProvider
	settings  settings.Repository
	collector TickerCollector
	emitter   EventEmitter
	snapshots stock.SnapshotRepository
	alerts    alert.Repository

	mu               sync.Mutex
	status           RefreshStatus
	running          atomic.Bool
	cancel           context.CancelFunc
	done             chan struct{}
	intervalOverride int // minutes, 0 = no override
}

// NewRefreshService creates a new RefreshService.
func NewRefreshService(
	stockData stock.Repository,
	provider stock.DataProvider,
	settings settings.Repository,
	collector TickerCollector,
	emitter EventEmitter,
	snapshots stock.SnapshotRepository,
	alerts alert.Repository,
) *RefreshService {
	return &RefreshService{
		stockData: stockData,
		provider:  provider,
		settings:  settings,
		collector: collector,
		emitter:   emitter,
		snapshots: snapshots,
		alerts:    alerts,
		status:    RefreshStatus{State: stateIdle},
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

// SetIntervalOverride sets a temporary interval override in minutes.
// Pass 0 to clear the override.
func (r *RefreshService) SetIntervalOverride(minutes int) {
	r.mu.Lock()
	r.intervalOverride = minutes
	r.mu.Unlock()
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

	if err := r.refresh(ctx); err != nil {
		r.emitter.Emit(eventRefreshError, err.Error())
	}

	interval := r.readInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !r.shouldRefresh(ctx) {
				continue
			}
			if newInterval := r.effectiveInterval(ctx); newInterval > 0 && newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
			}
			if err := r.refresh(ctx); err != nil {
				r.emitter.Emit(eventRefreshError, err.Error())
			}
		}
	}
}

func (r *RefreshService) shouldRefresh(ctx context.Context) bool {
	cfg, err := r.settings.GetRefreshSettings(ctx)
	return err == nil && cfg.AutoRefreshEnabled
}

func (r *RefreshService) effectiveInterval(ctx context.Context) time.Duration {
	cfg, err := r.settings.GetRefreshSettings(ctx)
	if err != nil {
		return 0
	}
	interval := time.Duration(cfg.IntervalMinutes) * time.Minute
	r.mu.Lock()
	override := r.intervalOverride
	r.mu.Unlock()
	if override > 0 {
		if ov := time.Duration(override) * time.Minute; ov < interval {
			interval = ov
		}
	}
	return interval
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

	r.setStatus(RefreshStatus{State: stateSyncing})

	start := time.Now()

	tickers, err := r.collector.CollectAll(ctx)
	if err != nil {
		errStatus := RefreshStatus{State: stateError, Error: fmt.Sprintf("collect tickers: %v", err)}
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
	r.emitter.Emit(eventRefreshSummary, RefreshSummary{
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
		if saveErr := r.settings.SaveRefreshSettings(ctx, cfg); saveErr != nil {
			r.emitter.Emit(eventRefreshError, saveErr.Error())
		}
	}

	finalState := stateIdle
	if total > 0 && failed == total {
		finalState = stateError
	}
	r.setStatus(RefreshStatus{
		State:       finalState,
		LastRefresh: now.Format("2006-01-02T15:04:05Z"),
	})

	r.emitAlertCount(ctx)

	return nil
}

func (r *RefreshService) emitAlertCount(ctx context.Context) {
	if r.alerts == nil {
		return
	}
	count, err := r.alerts.CountActive(ctx)
	if err != nil {
		return
	}
	r.emitter.Emit("alerts:updated", count)
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

	r.detectAndStoreAlerts(ctx, data)

	return r.stockData.Upsert(ctx, data)
}

// detectAndStoreAlerts compares the current data with the previous snapshot,
// detects fundamental changes, and persists alerts. Errors are non-fatal.
func (r *RefreshService) detectAndStoreAlerts(ctx context.Context, data *stock.Data) {
	if r.snapshots == nil || r.alerts == nil {
		return
	}

	prev, err := r.snapshots.GetLatest(ctx, data.Ticker, data.Source)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return
	}

	// Store current as new snapshot.
	if insertErr := r.snapshots.Insert(ctx, data); insertErr != nil {
		return
	}

	// Detect changes if we have a previous snapshot.
	if prev != nil {
		detected := alert.DetectChanges(prev, data)
		r.reconcileAlerts(ctx, data.Ticker, detected)
	}

	// Cleanup old snapshots.
	_ = r.snapshots.Cleanup(ctx, data.Ticker, snapshotKeepN)
}

// reconcileAlerts creates new alerts for newly detected changes and
// auto-resolves existing alerts whose metrics have recovered.
// Uses a single GetActiveByTicker query to avoid duplicate DB calls.
func (r *RefreshService) reconcileAlerts(ctx context.Context, ticker string, detected []*alert.FundamentalAlert) {
	existing, err := r.alerts.GetActiveByTicker(ctx, ticker)
	if err != nil {
		return
	}

	activeMetrics := make(map[string]bool, len(existing))
	for _, a := range existing {
		activeMetrics[a.Metric] = true
	}

	detectedMetrics := make(map[string]bool, len(detected))
	for _, d := range detected {
		detectedMetrics[d.Metric] = true
		if !activeMetrics[d.Metric] {
			_ = r.alerts.Create(ctx, d)
		}
	}

	for _, a := range existing {
		if !detectedMetrics[a.Metric] {
			_ = r.alerts.Resolve(ctx, a.ID)
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
