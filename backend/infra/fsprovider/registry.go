package fsprovider

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
	domainProvider "github.com/lugassawan/panen/backend/domain/provider"
	"github.com/lugassawan/panen/backend/infra/applog"
)

const (
	sourceRegistry    = "fs-registry"
	healthCheckTicker = "BBCA"
)

// entry holds a provider alongside its registration metadata.
type entry struct {
	provider  financials.FinancialStatementProvider
	priority  int
	status    domainProvider.Status
	lastCheck time.Time
	lastError string
	enabled   bool
}

// Registry manages multiple FinancialStatementProvider implementations with
// priority ordering and automatic fallback.
type Registry struct {
	mu      sync.RWMutex
	entries []entry
}

// Compile-time checks.
var (
	_ financials.FinancialStatementProvider = (*Registry)(nil)
	_ financials.Registry                   = (*Registry)(nil)
)

// NewRegistry creates an empty financial statement provider registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a provider with the given priority (lower = higher priority).
func (r *Registry) Register(p financials.FinancialStatementProvider, priority int) {
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
func (r *Registry) Source() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, e := range r.entries {
		if e.enabled {
			return e.provider.Source()
		}
	}
	return sourceRegistry
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

	if !enabled && enabledCount <= 1 && r.entries[idx].enabled {
		return false
	}

	r.entries[idx].enabled = enabled

	if !enabled {
		r.entries[idx].status = domainProvider.StatusUnknown
		r.entries[idx].lastError = ""
	}

	return true
}

// FetchIncomeStatements tries each enabled provider in priority order until one succeeds.
func (r *Registry) FetchIncomeStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	var lastErr error
	for _, e := range r.enabledEntries() {
		result, err := e.provider.FetchIncomeStatements(ctx, ticker, period)
		if err == nil {
			return result, nil
		}
		lastErr = err
		r.logFallback("FetchIncomeStatements", e.provider.Source(), ticker, err)
	}
	if lastErr == nil {
		return nil, financials.ErrNoStatements
	}
	return nil, lastErr
}

// FetchBalanceSheets tries each enabled provider in priority order until one succeeds.
func (r *Registry) FetchBalanceSheets(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	var lastErr error
	for _, e := range r.enabledEntries() {
		result, err := e.provider.FetchBalanceSheets(ctx, ticker, period)
		if err == nil {
			return result, nil
		}
		lastErr = err
		r.logFallback("FetchBalanceSheets", e.provider.Source(), ticker, err)
	}
	if lastErr == nil {
		return nil, financials.ErrNoStatements
	}
	return nil, lastErr
}

// FetchCashFlowStatements tries each enabled provider in priority order until one succeeds.
func (r *Registry) FetchCashFlowStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	var lastErr error
	for _, e := range r.enabledEntries() {
		result, err := e.provider.FetchCashFlowStatements(ctx, ticker, period)
		if err == nil {
			return result, nil
		}
		lastErr = err
		r.logFallback("FetchCashFlowStatements", e.provider.Source(), ticker, err)
	}
	if lastErr == nil {
		return nil, financials.ErrNoStatements
	}
	return nil, lastErr
}

// HealthCheckAll runs health checks on all registered providers.
func (r *Registry) HealthCheckAll(ctx context.Context) {
	snapshot := r.enabledEntries()

	for _, e := range snapshot {
		status := domainProvider.StatusHealthy
		var errMsg string

		_, err := e.provider.FetchIncomeStatements(ctx, healthCheckTicker, financials.PeriodAnnual)
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
	applog.Warn("fs provider fallback", err, applog.Fields{
		"method":   method,
		"provider": source,
		"ticker":   ticker,
	})
}
