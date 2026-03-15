package fmp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lugassawan/panen/backend/domain/financials"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) *FMP {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewFMP("test-key", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
}

func TestFMPSource(t *testing.T) {
	f := NewFMP("key")
	if got := f.Source(); got != "fmp" {
		t.Errorf("Source() = %q, want %q", got, "fmp")
	}
}

func TestFMPFetchIncomeStatements(t *testing.T) {
	resp := []incomeStatementResponse{
		{
			CalendarYear: "2024",
			Period:       "FY",
			Revenue:      50000000,
			NetIncome:    15000000,
			EPS:          500,
			EPSDiluted:   495,
			EBITDA:       20000000,
		},
		{
			CalendarYear: "2023",
			Period:       "FY",
			Revenue:      45000000,
			NetIncome:    13000000,
			EPS:          450,
		},
	}

	f := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/income-statement/BBCA.JK" {
			t.Errorf("request path = %q, want /income-statement/BBCA.JK", got)
		}
		if got := r.URL.Query().Get("period"); got != "annual" {
			t.Errorf("period = %q, want annual", got)
		}
		if got := r.URL.Query().Get("apikey"); got != "test-key" {
			t.Errorf("apikey = %q, want test-key", got)
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	stmts, err := f.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchIncomeStatements() error = %v", err)
	}
	if len(stmts) != 2 {
		t.Fatalf("len(stmts) = %d, want 2", len(stmts))
	}
	if stmts[0].Revenue != 50000000 {
		t.Errorf("stmts[0].Revenue = %v, want 50000000", stmts[0].Revenue)
	}
	if stmts[0].FiscalYear != 2024 {
		t.Errorf("stmts[0].FiscalYear = %d, want 2024", stmts[0].FiscalYear)
	}
	if stmts[0].Quarter != 0 {
		t.Errorf("stmts[0].Quarter = %d, want 0", stmts[0].Quarter)
	}
	if stmts[0].Source != "fmp" {
		t.Errorf("stmts[0].Source = %q, want fmp", stmts[0].Source)
	}
	if stmts[0].Ticker != "BBCA" {
		t.Errorf("stmts[0].Ticker = %q, want BBCA", stmts[0].Ticker)
	}
}

func TestFMPFetchIncomeStatementsQuarterly(t *testing.T) {
	resp := []incomeStatementResponse{
		{CalendarYear: "2024", Period: "Q3", Revenue: 13000000},
		{CalendarYear: "2024", Period: "Q2", Revenue: 12500000},
		{CalendarYear: "2024", Period: "Q1", Revenue: 12000000},
	}

	f := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("period"); got != "quarterly" {
			t.Errorf("period = %q, want quarterly", got)
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	stmts, err := f.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodQuarterly)
	if err != nil {
		t.Fatalf("FetchIncomeStatements() error = %v", err)
	}
	if len(stmts) != 3 {
		t.Fatalf("len(stmts) = %d, want 3", len(stmts))
	}
	if stmts[0].Quarter != 3 {
		t.Errorf("stmts[0].Quarter = %d, want 3", stmts[0].Quarter)
	}
}

func TestFMPFetchBalanceSheets(t *testing.T) {
	resp := []balanceSheetResponse{
		{
			CalendarYear:            "2024",
			Period:                  "FY",
			TotalAssets:             1000000000,
			TotalStockholdersEquity: 300000000,
			TotalDebt:               400000000,
		},
	}

	f := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/balance-sheet-statement/BBCA.JK" {
			t.Errorf("request path = %q, want /balance-sheet-statement/BBCA.JK", got)
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	sheets, err := f.FetchBalanceSheets(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchBalanceSheets() error = %v", err)
	}
	if len(sheets) != 1 {
		t.Fatalf("len(sheets) = %d, want 1", len(sheets))
	}
	if sheets[0].TotalAssets != 1000000000 {
		t.Errorf("TotalAssets = %v, want 1000000000", sheets[0].TotalAssets)
	}
	if sheets[0].TotalEquity != 300000000 {
		t.Errorf("TotalEquity = %v, want 300000000", sheets[0].TotalEquity)
	}
	if sheets[0].TotalDebt != 400000000 {
		t.Errorf("TotalDebt = %v, want 400000000", sheets[0].TotalDebt)
	}
	if sheets[0].Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want BBCA", sheets[0].Ticker)
	}
}

func TestFMPFetchCashFlowStatements(t *testing.T) {
	resp := []cashFlowResponse{
		{
			CalendarYear:      "2024",
			Period:            "FY",
			OperatingCashFlow: 20000000,
			FreeCashFlow:      15000000,
			DividendsPaid:     5000000,
		},
	}

	f := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/cash-flow-statement/BBCA.JK" {
			t.Errorf("request path = %q, want /cash-flow-statement/BBCA.JK", got)
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	stmts, err := f.FetchCashFlowStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchCashFlowStatements() error = %v", err)
	}
	if len(stmts) != 1 {
		t.Fatalf("len(stmts) = %d, want 1", len(stmts))
	}
	if stmts[0].OperatingCashFlow != 20000000 {
		t.Errorf("OperatingCashFlow = %v, want 20000000", stmts[0].OperatingCashFlow)
	}
	if stmts[0].FreeCashFlow != 15000000 {
		t.Errorf("FreeCashFlow = %v, want 15000000", stmts[0].FreeCashFlow)
	}
}

func TestFMPHTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{"forbidden returns ErrInvalidKey", http.StatusForbidden, financials.ErrInvalidKey},
		{"rate limited returns ErrRateLimited", http.StatusTooManyRequests, financials.ErrRateLimited},
		{"server error returns ErrSourceDown", http.StatusInternalServerError, financials.ErrSourceDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			_, err := f.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestFMPEmptyResponse(t *testing.T) {
	f := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]incomeStatementResponse{})
	})

	_, err := f.FetchIncomeStatements(context.Background(), "BBCA", financials.PeriodAnnual)
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}

func TestFMPTickerWithJKSuffix(t *testing.T) {
	f := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/income-statement/BBCA.JK" {
			t.Errorf("request path = %q, want /income-statement/BBCA.JK", got)
		}
		_ = json.NewEncoder(w).Encode([]incomeStatementResponse{
			{CalendarYear: "2024", Period: "FY", Revenue: 1000},
		})
	})

	// Passing ticker with .JK suffix should not double-append
	_, err := f.FetchIncomeStatements(context.Background(), "BBCA.JK", financials.PeriodAnnual)
	if err != nil {
		t.Fatalf("FetchIncomeStatements() error = %v", err)
	}
}

func TestParseFMPPeriod(t *testing.T) {
	tests := []struct {
		calYear string
		period  string
		wantY   int
		wantQ   int
	}{
		{"2024", "FY", 2024, 0},
		{"2024", "Q1", 2024, 1},
		{"2023", "Q4", 2023, 4},
		{"", "FY", 0, 0},
	}

	for _, tt := range tests {
		y, q := parseFMPPeriod(tt.calYear, tt.period)
		if y != tt.wantY || q != tt.wantQ {
			t.Errorf("parseFMPPeriod(%q, %q) = (%d, %d), want (%d, %d)",
				tt.calYear, tt.period, y, q, tt.wantY, tt.wantQ)
		}
	}
}
