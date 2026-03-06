package presenter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
	"github.com/lugassawan/panen/backend/usecase"
)

// Compile-time interface checks.
var (
	_ stock.Repository   = (*mockStockRepo)(nil)
	_ stock.DataProvider = (*mockDataProvider)(nil)
)

// --- mock stock data provider ---

type mockDataProvider struct {
	source    string
	priceFunc func(ctx context.Context, ticker string) (*stock.PriceResult, error)
	finFunc   func(ctx context.Context, ticker string) (*stock.FinancialResult, error)
}

func (m *mockDataProvider) Source() string { return m.source }

func (m *mockDataProvider) FetchPrice(ctx context.Context, ticker string) (*stock.PriceResult, error) {
	if m.priceFunc != nil {
		return m.priceFunc(ctx, ticker)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDataProvider) FetchFinancials(ctx context.Context, ticker string) (*stock.FinancialResult, error) {
	if m.finFunc != nil {
		return m.finFunc(ctx, ticker)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDataProvider) FetchPriceHistory(_ context.Context, _ string) ([]stock.PricePoint, error) {
	return nil, nil
}

func (m *mockDataProvider) FetchDividendHistory(_ context.Context, _ string) ([]dividend.DividendEvent, error) {
	return nil, nil
}

// --- helper to build a stock handler with mocks ---

func newTestStockHandler(provider *mockDataProvider) (*StockHandler, *mockStockRepo) {
	stockRepo := newMockStockRepo()
	svc := usecase.NewStockService(stockRepo, provider)
	ctx := context.Background()
	handler := NewStockHandler(ctx, svc)
	return handler, stockRepo
}

// --- test helpers ---

func checkField(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", field, got, want)
	}
}

func checkNonZero(t *testing.T, field string, got any) {
	t.Helper()
	if got == "" || got == 0 || got == 0.0 {
		t.Errorf("%s should not be zero/empty", field)
	}
}

// --- tests ---

func TestLookupStock(t *testing.T) {
	t.Run("fetches from provider when no cache", func(t *testing.T) {
		provider := &mockDataProvider{
			source: "test",
			priceFunc: func(_ context.Context, _ string) (*stock.PriceResult, error) {
				return &stock.PriceResult{
					Price:      9500,
					High52Week: 10000,
					Low52Week:  8000,
				}, nil
			},
			finFunc: func(_ context.Context, _ string) (*stock.FinancialResult, error) {
				return &stock.FinancialResult{
					EPS:           500,
					BVPS:          4000,
					ROE:           12.5,
					DER:           0.8,
					PBV:           2.375,
					PER:           19,
					DividendYield: 2.5,
					PayoutRatio:   40,
				}, nil
			},
		}
		handler, _ := newTestStockHandler(provider)

		resp, err := handler.LookupStock("BBCA", "MODERATE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		checkField(t, "Ticker", resp.Ticker, "BBCA")
		checkField(t, "Price", resp.Price, 9500.0)
		checkField(t, "High52Week", resp.High52Week, 10000.0)
		checkField(t, "Low52Week", resp.Low52Week, 8000.0)
		checkField(t, "EPS", resp.EPS, 500.0)
		checkField(t, "BVPS", resp.BVPS, 4000.0)
		checkField(t, "ROE", resp.ROE, 12.5)
		checkField(t, "DER", resp.DER, 0.8)
		checkField(t, "PBV", resp.PBV, 2.375)
		checkField(t, "PER", resp.PER, 19.0)
		checkField(t, "DividendYield", resp.DividendYield, 2.5)
		checkField(t, "PayoutRatio", resp.PayoutRatio, 40.0)
		checkField(t, "RiskProfile", resp.RiskProfile, "MODERATE")
		checkField(t, "Source", resp.Source, "test")
		checkNonZero(t, "GrahamNumber", resp.GrahamNumber)
		checkNonZero(t, "Verdict", resp.Verdict)
		checkNonZero(t, "FetchedAt", resp.FetchedAt)
	})

	t.Run("uses cache when fresh", func(t *testing.T) {
		fetchCalled := false
		provider := &mockDataProvider{
			source: "test",
			priceFunc: func(_ context.Context, _ string) (*stock.PriceResult, error) {
				fetchCalled = true
				return &stock.PriceResult{Price: 9500}, nil
			},
			finFunc: func(_ context.Context, _ string) (*stock.FinancialResult, error) {
				fetchCalled = true
				return &stock.FinancialResult{EPS: 500, BVPS: 4000}, nil
			},
		}
		handler, stockRepo := newTestStockHandler(provider)

		// Pre-populate cache with fresh data
		stockRepo.data["BBCA:test"] = &stock.Data{
			ID:        "s1",
			Ticker:    "BBCA",
			Price:     9000,
			EPS:       500,
			BVPS:      4000,
			PBV:       2.25,
			PER:       18,
			FetchedAt: time.Now().UTC(), // fresh cache
			Source:    "test",
		}

		resp, err := handler.LookupStock("BBCA", "MODERATE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fetchCalled {
			t.Error("expected to use cache, but provider was called")
		}
		if resp.Price != 9000 {
			t.Errorf("Price = %v, want 9000 (from cache)", resp.Price)
		}
	})

	t.Run("invalid risk profile", func(t *testing.T) {
		provider := &mockDataProvider{source: "test"}
		handler, _ := newTestStockHandler(provider)

		_, err := handler.LookupStock("BBCA", "INVALID")
		if err == nil {
			t.Fatal("expected error for invalid risk profile")
		}
		if !errors.Is(err, valuation.ErrInvalidRisk) {
			t.Errorf("error = %v, want ErrInvalidRisk", err)
		}
	})
}

func TestGetStockValuation(t *testing.T) {
	t.Run("returns cached valuation", func(t *testing.T) {
		provider := &mockDataProvider{source: "test"}
		handler, stockRepo := newTestStockHandler(provider)

		stockRepo.data["BBCA:test"] = &stock.Data{
			ID:        "s1",
			Ticker:    "BBCA",
			Price:     9500,
			EPS:       500,
			BVPS:      4000,
			ROE:       12.5,
			DER:       0.8,
			PBV:       2.375,
			PER:       19,
			FetchedAt: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			Source:    "test",
		}

		resp, err := handler.GetStockValuation("BBCA", "CONSERVATIVE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		checkField(t, "Ticker", resp.Ticker, "BBCA")
		checkField(t, "Price", resp.Price, 9500.0)
		checkField(t, "RiskProfile", resp.RiskProfile, "CONSERVATIVE")
		checkNonZero(t, "GrahamNumber", resp.GrahamNumber)
		checkNonZero(t, "Verdict", resp.Verdict)
	})

	t.Run("no cached data", func(t *testing.T) {
		provider := &mockDataProvider{source: "test"}
		handler, _ := newTestStockHandler(provider)

		_, err := handler.GetStockValuation("XXXX", "MODERATE")
		if err == nil {
			t.Fatal("expected error for no cached data")
		}
		if !errors.Is(err, usecase.ErrNoStockData) {
			t.Errorf("error = %v, want ErrNoStockData", err)
		}
	})

	t.Run("invalid risk profile", func(t *testing.T) {
		provider := &mockDataProvider{source: "test"}
		handler, _ := newTestStockHandler(provider)

		_, err := handler.GetStockValuation("BBCA", "INVALID")
		if err == nil {
			t.Fatal("expected error for invalid risk profile")
		}
		if !errors.Is(err, valuation.ErrInvalidRisk) {
			t.Errorf("error = %v, want ErrInvalidRisk", err)
		}
	})

	t.Run("response fields match stock data", func(t *testing.T) {
		provider := &mockDataProvider{source: "test"}
		handler, stockRepo := newTestStockHandler(provider)

		fetchedAt := time.Date(2025, 7, 15, 14, 30, 0, 0, time.UTC)
		stockRepo.data["BBRI:test"] = &stock.Data{
			ID:            "s2",
			Ticker:        "BBRI",
			Price:         5000,
			High52Week:    5500,
			Low52Week:     4000,
			EPS:           300,
			BVPS:          2500,
			ROE:           12.0,
			DER:           1.5,
			PBV:           2.0,
			PER:           16.67,
			DividendYield: 3.0,
			PayoutRatio:   50,
			FetchedAt:     fetchedAt,
			Source:        "test",
		}

		resp, err := handler.GetStockValuation("BBRI", "MODERATE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		checkField(t, "High52Week", resp.High52Week, 5500.0)
		checkField(t, "Low52Week", resp.Low52Week, 4000.0)
		checkField(t, "DividendYield", resp.DividendYield, 3.0)
		checkField(t, "PayoutRatio", resp.PayoutRatio, 50.0)
		checkField(t, "FetchedAt", resp.FetchedAt, "2025-07-15T14:30:00Z")
		checkField(t, "Source", resp.Source, "test")
		checkField(t, "RiskProfile", resp.RiskProfile, "MODERATE")
	})
}

func TestLookupStockProviderError(t *testing.T) {
	providerErr := errors.New("network error")
	provider := &mockDataProvider{
		source: "test",
		priceFunc: func(_ context.Context, _ string) (*stock.PriceResult, error) {
			return nil, providerErr
		},
	}
	handler, _ := newTestStockHandler(provider)

	_, err := handler.LookupStock("BBCA", "MODERATE")
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("error = %v, want %v", err, providerErr)
	}
}

func TestLookupStockNormalizesTickerCase(t *testing.T) {
	provider := &mockDataProvider{
		source: "test",
		priceFunc: func(_ context.Context, ticker string) (*stock.PriceResult, error) {
			if ticker != "BBCA" {
				t.Errorf("ticker = %q, want BBCA (normalized to uppercase)", ticker)
			}
			return &stock.PriceResult{Price: 9500}, nil
		},
		finFunc: func(_ context.Context, _ string) (*stock.FinancialResult, error) {
			return &stock.FinancialResult{EPS: 500, BVPS: 4000}, nil
		},
	}
	handler, _ := newTestStockHandler(provider)

	resp, err := handler.LookupStock("bbca", "MODERATE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want BBCA", resp.Ticker)
	}
}

func TestGetStockValuationTickerNotCached(t *testing.T) {
	provider := &mockDataProvider{source: "test"}
	handler, _ := newTestStockHandler(provider)

	_, err := handler.GetStockValuation("NONEXISTENT", "MODERATE")
	if err == nil {
		t.Fatal("expected error for uncached ticker")
	}
	if !errors.Is(err, usecase.ErrNoStockData) {
		t.Errorf("error = %v, want ErrNoStockData", err)
	}
}
