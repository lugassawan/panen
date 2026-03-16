package dataset

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFetchFinancials(t *testing.T) {
	body := `{
		"version": 1,
		"updatedAt": "2026-03-15",
		"data": [
			{"year": 2025, "quarter": 4, "metric": "revenue", "value": 1000000},
			{"year": 2025, "quarter": 4, "metric": "net_income", "value": 200000},
			{"year": 2025, "quarter": 3, "metric": "total_equity", "value": 500000}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "financials" {
			t.Errorf("expected action=financials, got %s", r.URL.Query().Get("action"))
		}
		if r.URL.Query().Get("ticker") != "BBCA" {
			t.Errorf("expected ticker=BBCA, got %s", r.URL.Query().Get("ticker"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	result, err := client.FetchFinancials(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Revenue) != 1 {
		t.Fatalf("expected 1 revenue entry, got %d", len(result.Revenue))
	}
	if result.Revenue[0].Value != 1000000 {
		t.Errorf("expected revenue value 1000000, got %f", result.Revenue[0].Value)
	}
	if len(result.NetIncome) != 1 {
		t.Fatalf("expected 1 net_income entry, got %d", len(result.NetIncome))
	}
	if result.NetIncome[0].Year != 2025 || result.NetIncome[0].Quarter != 4 {
		t.Errorf("unexpected net_income entry: %+v", result.NetIncome[0])
	}
	if len(result.TotalEquity) != 1 {
		t.Fatalf("expected 1 total_equity entry, got %d", len(result.TotalEquity))
	}
}

func TestFetchFinancialsCacheHit(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(
			[]byte(
				`{"version":1,"updatedAt":"2026-03-15","data":[{"year":2025,"quarter":4,"metric":"revenue","value":999}]}`,
			),
		)
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	// First call — hits server.
	_, err := client.FetchFinancials(context.Background(), "BMRI")
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 server call, got %d", callCount)
	}

	// Second call — should hit cache.
	result, err := client.FetchFinancials(context.Background(), "BMRI")
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected cache hit (1 server call), got %d", callCount)
	}
	if result == nil || len(result.Revenue) != 1 || result.Revenue[0].Value != 999 {
		t.Errorf("unexpected cached result: %+v", result)
	}
}

func TestFetchFinancialsCacheExpired(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(
			[]byte(
				`{"version":1,"updatedAt":"2026-03-15","data":[{"year":2025,"quarter":4,"metric":"revenue","value":42}]}`,
			),
		)
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	// First call.
	_, err := client.FetchFinancials(context.Background(), "TLKM")
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}

	// Backdate the cache file to simulate expiry.
	cachePath := filepath.Join(dataDir, "datasets", "financials_TLKM.json")
	expired := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(cachePath, expired, expired); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	// Second call — cache is expired, should hit server again.
	_, err = client.FetchFinancials(context.Background(), "TLKM")
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 server calls after expiry, got %d", callCount)
	}
}

func TestEmptyURLGracefulDegradation(t *testing.T) {
	dataDir := t.TempDir()
	client := NewClient("", dataDir)

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "FetchFinancials",
			fn: func() error {
				result, err := client.FetchFinancials(context.Background(), "BBCA")
				if result != nil {
					t.Errorf("expected nil result for FetchFinancials")
				}
				return err
			},
		},
		{
			name: "FetchAllFinancials",
			fn: func() error {
				result, err := client.FetchAllFinancials(context.Background())
				if result != nil {
					t.Errorf("expected nil result for FetchAllFinancials")
				}
				return err
			},
		},
		{
			name: "FetchSectorMetrics",
			fn: func() error {
				result, err := client.FetchSectorMetrics(context.Background(), "BBCA")
				if result != nil {
					t.Errorf("expected nil result for FetchSectorMetrics")
				}
				return err
			},
		},
		{
			name: "FetchDefinitions",
			fn: func() error {
				result, err := client.FetchDefinitions(context.Background())
				if result != nil {
					t.Errorf("expected nil result for FetchDefinitions")
				}
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestResponseSizeLimit(t *testing.T) {
	// Create a response larger than 1 MB.
	largeBody := make([]byte, maxResponseBytes+1)
	for i := range largeBody {
		largeBody[i] = 'x'
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(largeBody)
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	_, err := client.FetchFinancials(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("expected error for oversized response")
	}
}

func TestMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	_, err := client.FetchFinancials(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestFetchSectorMetrics(t *testing.T) {
	body := `{
		"version": 1,
		"updatedAt": "2026-03-15",
		"data": {
			"nim": [
				{"year": 2025, "quarter": 4, "value": 5.2, "source": "auto"},
				{"year": 2025, "quarter": 3, "value": 5.1, "source": "manual"}
			]
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	result, err := client.FetchSectorMetrics(context.Background(), "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	nim, ok := result["nim"]
	if !ok {
		t.Fatal("expected 'nim' metric")
	}
	if len(nim) != 2 {
		t.Fatalf("expected 2 nim entries, got %d", len(nim))
	}
	if nim[0].Source != "auto" {
		t.Errorf("expected source 'auto', got %q", nim[0].Source)
	}
}

func TestFetchDefinitions(t *testing.T) {
	body := `{
		"version": 1,
		"updatedAt": "2026-03-15",
		"data": {
			"Banking": ["nim", "car", "ldr"],
			"Mining": ["stripping_ratio", "asc"]
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	result, err := client.FetchDefinitions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result["Banking"]) != 3 {
		t.Errorf("expected 3 Banking metrics, got %d", len(result["Banking"]))
	}
	if len(result["Mining"]) != 2 {
		t.Errorf("expected 2 Mining metrics, got %d", len(result["Mining"]))
	}
}

func TestHTTPErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dataDir := t.TempDir()
	client := NewClient(srv.URL, dataDir)

	_, err := client.FetchFinancials(context.Background(), "BBCA")
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}
