package scraper

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/lugassawan/panen/backend/domain/stock"
)

const priceHistoryFixture = `{
  "chart": {
    "result": [{
      "meta": {"regularMarketPrice": 9875.0},
      "timestamp": [1704153600, 1704240000, 1704326400],
      "indicators": {
        "quote": [{
          "open":   [9000.0, 9100.0, null],
          "high":   [9200.0, 9300.0, null],
          "low":    [8900.0, 9000.0, null],
          "close":  [9100.0, 9250.0, null],
          "volume": [100000, 150000, null]
        }]
      }
    }],
    "error": null
  }
}`

const chartFixture = `{
  "chart": {
    "result": [{
      "meta": {"regularMarketPrice": 9875.0},
      "indicators": {
        "quote": [{
          "high": [9500.0, 10200.0, 9800.0, 10100.0],
          "low":  [9100.0, 9300.0, 8900.0, 9200.0]
        }]
      }
    }],
    "error": null
  }
}`

const quoteSummaryFixture = `{
  "quoteSummary": {
    "result": [{
      "defaultKeyStatistics": {
        "trailingEps":  {"raw": 512.5},
        "bookValue":    {"raw": 2150.0},
        "priceToBook":  {"raw": 4.59},
        "trailingPE":   {"raw": 19.27}
      },
      "financialData": {
        "returnOnEquity": {"raw": 0.2085},
        "debtToEquity":   {"raw": 45.3}
      },
      "summaryDetail": {
        "dividendYield": {"raw": 0.0312},
        "payoutRatio":   {"raw": 0.601}
      }
    }],
    "error": null
  }
}`

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newTestServerWithHandler(t, nil)
}

func newTestServerWithHandler(
	t *testing.T,
	override func(w http.ResponseWriter, r *http.Request) bool,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if override != nil && override(w, r) {
			return
		}
		switch {
		case r.URL.Path == "/v1/test/getcrumb":
			_, _ = w.Write([]byte("testcrumb123"))
		case strings.Contains(r.URL.Path, "/v8/finance/chart/"):
			if r.URL.Query().Get("range") == "5y" {
				handleTestEndpoint(w, r, priceHistoryFixture)
			} else {
				handleTestEndpoint(w, r, chartFixture)
			}
		case strings.Contains(r.URL.Path, "/v10/finance/quoteSummary/"):
			handleTestEndpoint(w, r, quoteSummaryFixture)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func handleTestEndpoint(w http.ResponseWriter, r *http.Request, successBody string) {
	switch {
	case strings.Contains(r.URL.Path, "NOTFOUND.JK"):
		w.WriteHeader(http.StatusNotFound)
	case strings.Contains(r.URL.Path, "RATELIMIT.JK"):
		w.WriteHeader(http.StatusTooManyRequests)
	case strings.Contains(r.URL.Path, "SERVERERR.JK"):
		w.WriteHeader(http.StatusInternalServerError)
	case strings.Contains(r.URL.Path, "BADJSON.JK"):
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json`))
	case strings.Contains(r.URL.Path, "EMPTY.JK"):
		w.Header().Set("Content-Type", "application/json")
		// Determine the empty response based on which endpoint prefix is present.
		if strings.Contains(r.URL.Path, "/v8/") {
			_, _ = w.Write([]byte(`{"chart":{"result":[],"error":null}}`))
		} else {
			_, _ = w.Write([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(successBody))
	}
}

func newTestYahoo(t *testing.T, srv *httptest.Server) *Yahoo {
	t.Helper()
	return NewYahoo(
		WithBaseURL(srv.URL),
		WithCookieURL(srv.URL),
		WithRateLimit(100, 100),
	)
}

func TestYahooSource(t *testing.T) {
	y := NewYahoo()
	if got := y.Source(); got != "yahoo" {
		t.Errorf("Source() = %q, want %q", got, "yahoo")
	}
}

func TestFetchPrice(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	y := newTestYahoo(t, srv)

	tests := []struct {
		name      string
		ticker    string
		wantPrice float64
		wantHigh  float64
		wantLow   float64
		wantErr   error
	}{
		{
			name:      "happy path",
			ticker:    "BBCA",
			wantPrice: 9875.0,
			wantHigh:  10200.0,
			wantLow:   8900.0,
		},
		{
			name:      "already has .JK suffix",
			ticker:    "BBCA.JK",
			wantPrice: 9875.0,
			wantHigh:  10200.0,
			wantLow:   8900.0,
		},
		{
			name:    "invalid ticker 404",
			ticker:  "NOTFOUND",
			wantErr: stock.ErrInvalidTicker,
		},
		{
			name:    "rate limited 429",
			ticker:  "RATELIMIT",
			wantErr: stock.ErrRateLimited,
		},
		{
			name:    "server error 500",
			ticker:  "SERVERERR",
			wantErr: stock.ErrSourceDown,
		},
		{
			name:    "malformed JSON",
			ticker:  "BADJSON",
			wantErr: stock.ErrNoData,
		},
		{
			name:    "empty result",
			ticker:  "EMPTY",
			wantErr: stock.ErrNoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := y.FetchPrice(context.Background(), tt.ticker)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Price != tt.wantPrice {
				t.Errorf("Price = %v, want %v", result.Price, tt.wantPrice)
			}
			if result.High52Week != tt.wantHigh {
				t.Errorf("High52Week = %v, want %v", result.High52Week, tt.wantHigh)
			}
			if result.Low52Week != tt.wantLow {
				t.Errorf("Low52Week = %v, want %v", result.Low52Week, tt.wantLow)
			}
		})
	}
}

func TestFetchFinancials(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	y := newTestYahoo(t, srv)

	tests := []struct {
		name    string
		ticker  string
		wantErr error
		check   func(t *testing.T, r *stock.FinancialResult)
	}{
		{
			name:   "happy path",
			ticker: "BBCA",
			check: func(t *testing.T, r *stock.FinancialResult) {
				t.Helper()
				assertFloat(t, "EPS", r.EPS, 512.5)
				assertFloat(t, "BVPS", r.BVPS, 2150.0)
				assertFloat(t, "ROE", r.ROE, 20.85)
				assertFloat(t, "DER", r.DER, 45.3)
				assertFloat(t, "PBV", r.PBV, 4.59)
				assertFloat(t, "PER", r.PER, 19.27)
				assertFloat(t, "DividendYield", r.DividendYield, 3.12)
				assertFloat(t, "PayoutRatio", r.PayoutRatio, 60.1)
			},
		},
		{
			name:    "invalid ticker 404",
			ticker:  "NOTFOUND",
			wantErr: stock.ErrInvalidTicker,
		},
		{
			name:    "rate limited 429",
			ticker:  "RATELIMIT",
			wantErr: stock.ErrRateLimited,
		},
		{
			name:    "server error 500",
			ticker:  "SERVERERR",
			wantErr: stock.ErrSourceDown,
		},
		{
			name:    "malformed JSON",
			ticker:  "BADJSON",
			wantErr: stock.ErrNoData,
		},
		{
			name:    "empty result",
			ticker:  "EMPTY",
			wantErr: stock.ErrNoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := y.FetchFinancials(context.Background(), tt.ticker)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestFetchPriceContextCancellation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	y := newTestYahoo(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := y.FetchPrice(ctx, "BBCA")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestFetchFinancialsContextCancellation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	y := newTestYahoo(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := y.FetchFinancials(ctx, "BBCA")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestFetchPriceWithCrumb(t *testing.T) {
	var gotCrumb string
	srv := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) bool {
		if strings.Contains(r.URL.Path, "/v8/finance/chart/") {
			gotCrumb = r.URL.Query().Get("crumb")
		}
		return false // let default handler run
	})
	defer srv.Close()

	y := newTestYahoo(t, srv)
	_, err := y.FetchPrice(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCrumb != "testcrumb123" {
		t.Errorf("crumb = %q, want %q", gotCrumb, "testcrumb123")
	}
}

func TestCrumbRefreshOn401(t *testing.T) {
	var chartCalls atomic.Int32
	srv := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) bool {
		if !strings.Contains(r.URL.Path, "/v8/finance/chart/BBCA.JK") {
			return false
		}
		n := chartCalls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(chartFixture))
		return true
	})
	defer srv.Close()

	y := newTestYahoo(t, srv)
	result, err := y.FetchPrice(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Price != 9875.0 {
		t.Errorf("Price = %v, want 9875.0", result.Price)
	}
	if got := chartCalls.Load(); got != 2 {
		t.Errorf("chart calls = %d, want 2 (initial 401 + retry)", got)
	}
}

func TestPersistent401ReturnsError(t *testing.T) {
	srv := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) bool {
		if strings.Contains(r.URL.Path, "/v8/finance/chart/") {
			w.WriteHeader(http.StatusUnauthorized)
			return true
		}
		return false
	})
	defer srv.Close()

	y := newTestYahoo(t, srv)
	_, err := y.FetchPrice(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("expected error for persistent 401, got nil")
	}
	if !errors.Is(err, stock.ErrSourceDown) {
		t.Errorf("expected ErrSourceDown, got: %v", err)
	}
}

func TestFetchPriceHistory(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	y := newTestYahoo(t, srv)

	t.Run("happy path", func(t *testing.T) {
		points, err := y.FetchPriceHistory(context.Background(), "BBCA")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// null entries should be skipped, so 2 valid points
		if len(points) != 2 {
			t.Fatalf("len(points) = %d, want 2", len(points))
		}
		if points[0].Open != 9000 {
			t.Errorf("points[0].Open = %v, want 9000", points[0].Open)
		}
		if points[0].Close != 9100 {
			t.Errorf("points[0].Close = %v, want 9100", points[0].Close)
		}
		if points[0].Volume != 100000 {
			t.Errorf("points[0].Volume = %v, want 100000", points[0].Volume)
		}
		if points[1].Close != 9250 {
			t.Errorf("points[1].Close = %v, want 9250", points[1].Close)
		}
		if points[0].Ticker != "BBCA" {
			t.Errorf("Ticker = %q, want BBCA", points[0].Ticker)
		}
		if points[0].Source != "yahoo" {
			t.Errorf("Source = %q, want yahoo", points[0].Source)
		}
	})

	t.Run("invalid ticker 404", func(t *testing.T) {
		_, err := y.FetchPriceHistory(context.Background(), "NOTFOUND")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, stock.ErrInvalidTicker) {
			t.Errorf("expected ErrInvalidTicker, got: %v", err)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		_, err := y.FetchPriceHistory(context.Background(), "EMPTY")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, stock.ErrNoData) {
			t.Errorf("expected ErrNoData, got: %v", err)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		_, err := y.FetchPriceHistory(context.Background(), "BADJSON")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, stock.ErrNoData) {
			t.Errorf("expected ErrNoData, got: %v", err)
		}
	})
}

func assertFloat(t *testing.T, field string, got, want float64) {
	t.Helper()
	const epsilon = 0.01
	diff := got - want
	if diff < -epsilon || diff > epsilon {
		t.Errorf("%s = %v, want %v", field, got, want)
	}
}
