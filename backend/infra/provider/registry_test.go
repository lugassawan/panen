package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/lugassawan/panen/backend/domain/dividend"
	domainProvider "github.com/lugassawan/panen/backend/domain/provider"
	"github.com/lugassawan/panen/backend/domain/stock"
)

// mockProvider is a minimal stock.DataProvider for testing.
type mockProvider struct {
	name        string
	priceResult *stock.PriceResult
	priceErr    error
	finResult   *stock.FinancialResult
	finErr      error
	historyPts  []stock.PricePoint
	historyErr  error
	divEvents   []dividend.DividendEvent
	divErr      error
}

func (m *mockProvider) Source() string { return m.name }

func (m *mockProvider) FetchPrice(_ context.Context, _ string) (*stock.PriceResult, error) {
	return m.priceResult, m.priceErr
}

func (m *mockProvider) FetchFinancials(_ context.Context, _ string) (*stock.FinancialResult, error) {
	return m.finResult, m.finErr
}

func (m *mockProvider) FetchPriceHistory(_ context.Context, _ string) ([]stock.PricePoint, error) {
	return m.historyPts, m.historyErr
}

func (m *mockProvider) FetchDividendHistory(_ context.Context, _ string) ([]dividend.DividendEvent, error) {
	return m.divEvents, m.divErr
}

func TestRegistryRegisterAndPrimary(t *testing.T) {
	reg := NewRegistry()

	if got := reg.Primary(); got != nil {
		t.Fatalf("Primary() on empty registry = %v, want nil", got)
	}

	p1 := &mockProvider{name: "p1"}
	p2 := &mockProvider{name: "p2"}

	reg.Register(p2, 2)
	reg.Register(p1, 1)

	got := reg.Primary()
	if got == nil {
		t.Fatal("Primary() = nil, want p1")
	}
	if got.Source() != "p1" {
		t.Errorf("Primary().Source() = %q, want %q", got.Source(), "p1")
	}
}

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry()
	p := &mockProvider{name: "yahoo"}
	reg.Register(p, 1)

	if got := reg.Get("yahoo"); got == nil {
		t.Fatal("Get(yahoo) = nil, want provider")
	}
	if got := reg.Get("nonexistent"); got != nil {
		t.Fatalf("Get(nonexistent) = %v, want nil", got)
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockProvider{name: "yahoo"}, 1)
	reg.Register(&mockProvider{name: IDXSource}, 2)

	infos := reg.List()
	if len(infos) != 2 {
		t.Fatalf("List() len = %d, want 2", len(infos))
	}
	if infos[0].Name != "yahoo" {
		t.Errorf("List()[0].Name = %q, want %q", infos[0].Name, "yahoo")
	}
	if infos[0].Priority != 1 {
		t.Errorf("List()[0].Priority = %d, want 1", infos[0].Priority)
	}
	if infos[1].Name != IDXSource {
		t.Errorf("List()[1].Name = %q, want %q", infos[1].Name, IDXSource)
	}
}

func TestRegistrySetEnabled(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockProvider{name: "yahoo"}, 1)
	reg.Register(&mockProvider{name: IDXSource}, 2)

	if !reg.SetEnabled("yahoo", false) {
		t.Fatal("SetEnabled(yahoo, false) = false, want true")
	}

	// Primary should now be idx since yahoo is disabled.
	got := reg.Primary()
	if got == nil || got.Source() != IDXSource {
		t.Errorf("Primary() after disabling yahoo = %v, want idx", got)
	}

	if reg.SetEnabled("nonexistent", true) {
		t.Fatal("SetEnabled(nonexistent) = true, want false")
	}
}

func TestRegistrySource(t *testing.T) {
	reg := NewRegistry()

	if got := reg.Source(); got != "registry" {
		t.Errorf("Source() on empty = %q, want %q", got, "registry")
	}

	reg.Register(&mockProvider{name: "yahoo"}, 1)
	if got := reg.Source(); got != "yahoo" {
		t.Errorf("Source() with yahoo = %q, want %q", got, "yahoo")
	}
}

func TestRegistryFetchPriceFallback(t *testing.T) {
	reg := NewRegistry()

	failing := &mockProvider{
		name:     "primary",
		priceErr: errors.New("primary down"),
	}
	succeeding := &mockProvider{
		name: "secondary",
		priceResult: &stock.PriceResult{
			Price:      9000,
			High52Week: 10000,
			Low52Week:  8000,
		},
	}

	reg.Register(failing, 1)
	reg.Register(succeeding, 2)

	result, err := reg.FetchPrice(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("FetchPrice with fallback: unexpected error: %v", err)
	}
	if result.Price != 9000 {
		t.Errorf("FetchPrice Price = %v, want 9000", result.Price)
	}
}

func TestRegistryFetchPriceAllFail(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&mockProvider{
		name:     "p1",
		priceErr: stock.ErrSourceDown,
	}, 1)
	reg.Register(&mockProvider{
		name:     "p2",
		priceErr: stock.ErrRateLimited,
	}, 2)

	_, err := reg.FetchPrice(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("FetchPrice: expected error when all providers fail")
	}
	// Last error should be from p2 (the last tried).
	if !errors.Is(err, stock.ErrRateLimited) {
		t.Errorf("FetchPrice error = %v, want ErrRateLimited", err)
	}
}

func TestRegistryFetchPriceNoProviders(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.FetchPrice(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("FetchPrice: expected error with no providers")
	}
	if !errors.Is(err, stock.ErrNoData) {
		t.Errorf("FetchPrice error = %v, want ErrNoData", err)
	}
}

func TestRegistryFetchFinancialsFallback(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&mockProvider{
		name:   "primary",
		finErr: errors.New("primary down"),
	}, 1)
	reg.Register(&mockProvider{
		name:      "secondary",
		finResult: &stock.FinancialResult{EPS: 500},
	}, 2)

	result, err := reg.FetchFinancials(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EPS != 500 {
		t.Errorf("EPS = %v, want 500", result.EPS)
	}
}

func TestRegistryFetchPriceHistoryFallback(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&mockProvider{
		name:       "primary",
		historyErr: errors.New("primary down"),
	}, 1)
	reg.Register(&mockProvider{
		name:       "secondary",
		historyPts: []stock.PricePoint{{Close: 9000}},
	}, 2)

	result, err := reg.FetchPriceHistory(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Close != 9000 {
		t.Errorf("Close = %v, want 9000", result[0].Close)
	}
}

func TestRegistryFetchDividendHistoryFallback(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&mockProvider{
		name:   "primary",
		divErr: errors.New("primary down"),
	}, 1)
	reg.Register(&mockProvider{
		name:      "secondary",
		divEvents: []dividend.DividendEvent{{Amount: 100}},
	}, 2)

	result, err := reg.FetchDividendHistory(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Amount != 100 {
		t.Errorf("Amount = %v, want 100", result[0].Amount)
	}
}

func TestRegistryFetchSkipsDisabled(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&mockProvider{
		name:        "primary",
		priceResult: &stock.PriceResult{Price: 5000},
	}, 1)
	reg.Register(&mockProvider{
		name:        "secondary",
		priceResult: &stock.PriceResult{Price: 9000},
	}, 2)

	reg.SetEnabled("primary", false)

	result, err := reg.FetchPrice(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Price != 9000 {
		t.Errorf("Price = %v, want 9000 (from secondary)", result.Price)
	}
}

func TestRegistryHealthCheckAll(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&mockProvider{
		name:        "healthy",
		priceResult: &stock.PriceResult{Price: 9000},
	}, 1)
	reg.Register(&mockProvider{
		name:     "down",
		priceErr: errors.New("connection refused"),
	}, 2)

	reg.HealthCheckAll(context.Background())

	infos := reg.List()
	if len(infos) != 2 {
		t.Fatalf("List() len = %d, want 2", len(infos))
	}

	if infos[0].Status != domainProvider.StatusHealthy {
		t.Errorf("healthy provider status = %q, want %q", infos[0].Status, domainProvider.StatusHealthy)
	}
	if infos[0].LastCheck.IsZero() {
		t.Error("healthy provider LastCheck is zero")
	}

	if infos[1].Status != domainProvider.StatusDown {
		t.Errorf("down provider status = %q, want %q", infos[1].Status, domainProvider.StatusDown)
	}
	if infos[1].LastError == "" {
		t.Error("down provider LastError is empty")
	}
}

func TestRegistrySetEnabledPreventsDisablingAll(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockProvider{name: "yahoo"}, 1)
	reg.Register(&mockProvider{name: IDXSource}, 2)

	// Disable one — should succeed.
	if !reg.SetEnabled("yahoo", false) {
		t.Fatal("SetEnabled(yahoo, false) = false, want true")
	}

	// Try to disable the last one — should fail.
	if reg.SetEnabled(IDXSource, false) {
		t.Fatal("SetEnabled(idx, false) = true, want false (last enabled)")
	}

	// Re-enable yahoo, then disable idx — should succeed.
	reg.SetEnabled("yahoo", true)
	if !reg.SetEnabled(IDXSource, false) {
		t.Fatal("SetEnabled(idx, false) after re-enabling yahoo = false, want true")
	}
}

func TestRegistrySetEnabledClearsStatus(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockProvider{
		name:        "yahoo",
		priceResult: &stock.PriceResult{Price: 9000},
	}, 1)
	reg.Register(&mockProvider{name: IDXSource}, 2)

	// Run health check so yahoo has a status.
	reg.HealthCheckAll(context.Background())

	infos := reg.List()
	if infos[0].Status != domainProvider.StatusHealthy {
		t.Fatalf("before disable: status = %q, want healthy", infos[0].Status)
	}

	// Disable yahoo — status should reset to unknown.
	reg.SetEnabled("yahoo", false)
	infos = reg.List()
	if infos[0].Status != domainProvider.StatusUnknown {
		t.Errorf("after disable: status = %q, want unknown", infos[0].Status)
	}
	if infos[0].LastError != "" {
		t.Errorf("after disable: lastError = %q, want empty", infos[0].LastError)
	}
}

func TestRegistryPriorityOrdering(t *testing.T) {
	reg := NewRegistry()

	// Register out of order.
	reg.Register(&mockProvider{name: "c"}, 3)
	reg.Register(&mockProvider{name: "a"}, 1)
	reg.Register(&mockProvider{name: "b"}, 2)

	infos := reg.List()
	if infos[0].Name != "a" || infos[1].Name != "b" || infos[2].Name != "c" {
		t.Errorf("ordering = [%s, %s, %s], want [a, b, c]",
			infos[0].Name, infos[1].Name, infos[2].Name)
	}
}
