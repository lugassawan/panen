package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lugassawan/panen/backend/domain/stock"
)

const idxFixture = `{
  "LastPrice": 9875.0,
  "High52Week": 10500.0,
  "Low52Week": 8200.0,
  "PER": 19.5,
  "PBV": 4.3
}`

func newTestIDXServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		switch code {
		case "NOTFOUND":
			w.WriteHeader(http.StatusNotFound)
		case "RATELIMIT":
			w.WriteHeader(http.StatusTooManyRequests)
		case "SERVERERR":
			w.WriteHeader(http.StatusInternalServerError)
		case "BADJSON":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{invalid`))
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(idxFixture))
		}
	}))
}

func newTestIDXProvider(t *testing.T, srv *httptest.Server) *IDXProvider {
	t.Helper()
	return NewIDXProvider(
		WithIDXBaseURL(srv.URL),
		WithIDXRateLimit(100, 100),
	)
}

func TestIDXProviderSource(t *testing.T) {
	p := NewIDXProvider()
	if got := p.Source(); got != "idx" {
		t.Errorf("Source() = %q, want %q", got, "idx")
	}
}

func TestIDXFetchPrice(t *testing.T) {
	srv := newTestIDXServer(t)
	defer srv.Close()
	p := newTestIDXProvider(t, srv)

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
			wantHigh:  10500.0,
			wantLow:   8200.0,
		},
		{
			name:      "strips JK suffix",
			ticker:    "BBCA.JK",
			wantPrice: 9875.0,
			wantHigh:  10500.0,
			wantLow:   8200.0,
		},
		{
			name:    "not found",
			ticker:  "NOTFOUND",
			wantErr: stock.ErrInvalidTicker,
		},
		{
			name:    "rate limited",
			ticker:  "RATELIMIT",
			wantErr: stock.ErrRateLimited,
		},
		{
			name:    "server error",
			ticker:  "SERVERERR",
			wantErr: stock.ErrSourceDown,
		},
		{
			name:    "malformed JSON",
			ticker:  "BADJSON",
			wantErr: stock.ErrNoData,
		},
		{
			name:    "index ticker unsupported",
			ticker:  "^JKSE",
			wantErr: stock.ErrNoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.FetchPrice(context.Background(), tt.ticker)

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

func TestIDXFetchFinancials(t *testing.T) {
	srv := newTestIDXServer(t)
	defer srv.Close()
	p := newTestIDXProvider(t, srv)

	result, err := p.FetchFinancials(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PER != 19.5 {
		t.Errorf("PER = %v, want 19.5", result.PER)
	}
	if result.PBV != 4.3 {
		t.Errorf("PBV = %v, want 4.3", result.PBV)
	}
}

func TestIDXFetchPriceHistoryNotSupported(t *testing.T) {
	p := NewIDXProvider()
	_, err := p.FetchPriceHistory(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, stock.ErrNoData) {
		t.Errorf("error = %v, want ErrNoData", err)
	}
}

func TestIDXFetchDividendHistoryNotSupported(t *testing.T) {
	p := NewIDXProvider()
	_, err := p.FetchDividendHistory(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, stock.ErrNoData) {
		t.Errorf("error = %v, want ErrNoData", err)
	}
}

func TestIDXFetchPriceContextCancellation(t *testing.T) {
	srv := newTestIDXServer(t)
	defer srv.Close()
	p := newTestIDXProvider(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.FetchPrice(ctx, "BBCA")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
