package sectorsapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lugassawan/panen/backend/domain/financials"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) *SectorsApp {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewSectorsApp("test-key", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
}

func TestSectorsAppSource(t *testing.T) {
	s := NewSectorsApp("key")
	if got := s.Source(); got != "sectors" {
		t.Errorf("Source() = %q, want %q", got, "sectors")
	}
}

func TestSectorsAppFetchIncomeStatements(t *testing.T) {
	reports := []financialReportResponse{
		{
			Symbol:    "BBCA",
			Year:      2024,
			Quarter:   0,
			Revenue:   50000000,
			NetIncome: 15000000,
			EPS:       500,
		},
		{
			Symbol:  "BBCA",
			Year:    2023,
			Quarter: 0,
			Revenue: 45000000,
		},
	}

	s := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/company/financials/BBCA/" {
			t.Errorf("request path = %q, want /company/financials/BBCA/", got)
		}
		if got := r.URL.Query().Get("period"); got != "annual" {
			t.Errorf("period = %q, want annual", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want Bearer test-key", got)
		}
		_ = json.NewEncoder(w).Encode(reports)
	})

	stmts, err := s.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchIncomeStatements() error = %v", err)
	}
	if len(stmts) != 2 {
		t.Fatalf("len(stmts) = %d, want 2", len(stmts))
	}
	if stmts[0].Revenue != 50000000 {
		t.Errorf("stmts[0].Revenue = %v, want 50000000", stmts[0].Revenue)
	}
	if stmts[0].Source != "sectors" {
		t.Errorf("stmts[0].Source = %q, want sectors", stmts[0].Source)
	}
}

func TestSectorsAppFetchBalanceSheets(t *testing.T) {
	reports := []financialReportResponse{
		{Year: 2024, TotalAssets: 1000000000, TotalEquity: 300000000},
	}

	s := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(reports)
	})

	sheets, err := s.FetchBalanceSheets(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchBalanceSheets() error = %v", err)
	}
	if len(sheets) != 1 {
		t.Fatalf("len(sheets) = %d, want 1", len(sheets))
	}
	if sheets[0].TotalAssets != 1000000000 {
		t.Errorf("TotalAssets = %v, want 1000000000", sheets[0].TotalAssets)
	}
}

func TestSectorsAppFetchCashFlowStatements(t *testing.T) {
	reports := []financialReportResponse{
		{Year: 2024, OperatingCashFlow: 20000000, FreeCashFlow: 15000000},
	}

	s := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(reports)
	})

	stmts, err := s.FetchCashFlowStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchCashFlowStatements() error = %v", err)
	}
	if len(stmts) != 1 {
		t.Fatalf("len(stmts) = %d, want 1", len(stmts))
	}
	if stmts[0].FreeCashFlow != 15000000 {
		t.Errorf("FreeCashFlow = %v, want 15000000", stmts[0].FreeCashFlow)
	}
}

func TestSectorsAppHTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"unauthorized", http.StatusUnauthorized},
		{"forbidden", http.StatusForbidden},
		{"rate limited", http.StatusTooManyRequests},
		{"server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			_, err := s.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestSectorsAppEmptyResponse(t *testing.T) {
	s := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]financialReportResponse{})
	})

	_, err := s.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}
