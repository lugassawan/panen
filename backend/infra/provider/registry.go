package provider

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/lugassawan/panen/backend/domain/dividend"
	domainProvider "github.com/lugassawan/panen/backend/domain/provider"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/infra/applog"
)

// entry holds a provider alongside its registration metadata.
type entry struct {
	provider  stock.DataProvider
	priority  int
	status    domainProvider.Status
	lastCheck time.Time
	lastError string
	enabled   bool
}

// Registry manages multiple DataProvider implementations with priority ordering
// and automatic fallback. It implements stock.DataProvider so it can be used
// as a drop-in replacement wherever a single provider is expected.
type Registry struct {
	mu      sync.RWMutex
	entries []entry
}

// Compile-time check that Registry implements stock.DataProvider.
var _ stock.DataProvider = (*Registry)(nil)

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a provider with the given priority (lower = higher priority).
func (r *Registry) Register(p stock.DataProvider, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries = append(r.entries, entry{
		provider: p,
		priority: priority,
		status:   domainProvider.StatusUnknown,
		enabled:  true,
	})
	sort.Slice(r.entries, func(i, j int) bool {
		return r.entries[i].priority < r.entries[j].priority
	})
}

// Source returns the source identifier of the primary provider.
// If no providers are registered, returns "registry".
func (r *Registry) Source() string {
	if p := r.Primary(); p != nil {
		return p.Source()
	}
	return "registry"
}

// Primary returns the highest-priority enabled provider, or nil if none.
func (r *Registry) Primary() stock.DataProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, e := range r.entries {
		if e.enabled {
			return e.provider
		}
	}
	return nil
}

// Get returns the provider with the given source name, or nil if not found.
func (r *Registry) Get(name string) stock.DataProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, e := range r.entries {
		if e.provider.Source() == name {
			return e.provider
		}
	}
	return nil
}

// List returns metadata about all registered providers.
func (r *Registry) List() []domainProvider.Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]domainProvider.Info, len(r.entries))
	for i, e := range r.entries {
		infos[i] = domainProvider.Info{
			Name:      e.provider.Source(),
			Priority:  e.priority,
			Status:    e.status,
			LastCheck: e.lastCheck,
			LastError: e.lastError,
			Enabled:   e.enabled,
		}
	}
	return infos
}

// SetEnabled enables or disables a provider by name.
// Returns false if the provider is not found or if disabling would leave no enabled providers.
func (r *Registry) SetEnabled(name string, enabled bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := -1
	enabledCount := 0
	for i, e := range r.entries {
		if e.provider.Source() == name {
			idx = i
		}
		if e.enabled {
			enabledCount++
		}
	}

	if idx < 0 {
		return false
	}

	// Prevent disabling the last enabled provider.
	if !enabled && enabledCount <= 1 && r.entries[idx].enabled {
		return false
	}

	r.entries[idx].enabled = enabled

	// Clear health status when disabling so stale data isn't shown.
	if !enabled {
		r.entries[idx].status = domainProvider.StatusUnknown
		r.entries[idx].lastError = ""
	}

	return true
}

// FetchPrice tries each enabled provider in priority order until one succeeds.
func (r *Registry) FetchPrice(ctx context.Context, ticker string) (*stock.PriceResult, error) {
	var lastErr error
	for _, e := range r.enabledEntries() {
		result, err := e.provider.FetchPrice(ctx, ticker)
		if err == nil {
			return result, nil
		}
		lastErr = err
		r.logFallback("FetchPrice", e.provider.Source(), ticker, err)
	}
	if lastErr == nil {
		return nil, stock.ErrNoData
	}
	return nil, lastErr
}

// FetchFinancials tries each enabled provider in priority order until one succeeds.
func (r *Registry) FetchFinancials(ctx context.Context, ticker string) (*stock.FinancialResult, error) {
	var lastErr error
	for _, e := range r.enabledEntries() {
		result, err := e.provider.FetchFinancials(ctx, ticker)
		if err == nil {
			return result, nil
		}
		lastErr = err
		r.logFallback("FetchFinancials", e.provider.Source(), ticker, err)
	}
	if lastErr == nil {
		return nil, stock.ErrNoData
	}
	return nil, lastErr
}

// FetchPriceHistory tries each enabled provider in priority order until one succeeds.
func (r *Registry) FetchPriceHistory(ctx context.Context, ticker string) ([]stock.PricePoint, error) {
	var lastErr error
	for _, e := range r.enabledEntries() {
		result, err := e.provider.FetchPriceHistory(ctx, ticker)
		if err == nil {
			return result, nil
		}
		lastErr = err
		r.logFallback("FetchPriceHistory", e.provider.Source(), ticker, err)
	}
	if lastErr == nil {
		return nil, stock.ErrNoData
	}
	return nil, lastErr
}

// FetchDividendHistory tries each enabled provider in priority order until one succeeds.
func (r *Registry) FetchDividendHistory(ctx context.Context, ticker string) ([]dividend.DividendEvent, error) {
	var lastErr error
	for _, e := range r.enabledEntries() {
		result, err := e.provider.FetchDividendHistory(ctx, ticker)
		if err == nil {
			return result, nil
		}
		lastErr = err
		r.logFallback("FetchDividendHistory", e.provider.Source(), ticker, err)
	}
	if lastErr == nil {
		return nil, stock.ErrNoData
	}
	return nil, lastErr
}

// HealthCheckAll runs health checks on all registered providers and updates their status.
func (r *Registry) HealthCheckAll(ctx context.Context) {
	snapshot := r.enabledEntries()

	for _, e := range snapshot {
		status := domainProvider.StatusHealthy
		var errMsg string

		_, err := e.provider.FetchPrice(ctx, "BBCA")
		if err != nil {
			status = domainProvider.StatusDown
			errMsg = err.Error()
		}

		r.updateStatus(e.provider.Source(), status, errMsg)
	}
}

func (r *Registry) updateStatus(name string, status domainProvider.Status, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for i, e := range r.entries {
		if e.provider.Source() == name {
			r.entries[i].status = status
			r.entries[i].lastCheck = now
			r.entries[i].lastError = errMsg
			return
		}
	}
}

// enabledEntries returns a snapshot of enabled entries in priority order.
func (r *Registry) enabledEntries() []entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var enabled []entry
	for _, e := range r.entries {
		if e.enabled {
			enabled = append(enabled, e)
		}
	}
	return enabled
}

func (r *Registry) logFallback(method, source, ticker string, err error) {
	applog.Warn("provider fallback", err, applog.Fields{
		"method":   method,
		"provider": source,
		"ticker":   ticker,
	})
}
