package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/lugassawan/panen/backend/domain/financials"
)

const (
	defaultBaseURL  = "https://financialmodelingprep.com/api/v3"
	maxResponseSize = 5 << 20 // 5 MB
	sourceFMP       = "fmp"
)

// FMP fetches financial statement data from Financial Modeling Prep.
type FMP struct {
	client  *http.Client
	limiter *rate.Limiter
	baseURL string
	apiKey  string
}

// Option configures the FMP provider.
type Option func(*FMP)

// WithBaseURL overrides the base URL (useful for tests).
func WithBaseURL(u string) Option {
	return func(f *FMP) { f.baseURL = u }
}

// WithRateLimit sets custom rate limiter parameters.
func WithRateLimit(rps float64, burst int) Option {
	return func(f *FMP) { f.limiter = rate.NewLimiter(rate.Limit(rps), burst) }
}

// WithHTTPClient overrides the HTTP client (useful for tests).
func WithHTTPClient(c *http.Client) Option {
	return func(f *FMP) { f.client = c }
}

// NewFMP creates an FMP provider with the given API key and options.
func NewFMP(apiKey string, opts ...Option) *FMP {
	f := &FMP{
		limiter: rate.NewLimiter(4, 8),
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
	}
	for _, o := range opts {
		o(f)
	}
	if f.client == nil {
		f.client = &http.Client{}
	}
	return f
}

// Source returns the provider identifier.
func (f *FMP) Source() string { return sourceFMP }

// FetchIncomeStatements returns income statements for a ticker and period type.
func (f *FMP) FetchIncomeStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	fmpTicker := formatIDXTicker(ticker)
	reqURL := fmt.Sprintf("%s/income-statement/%s?period=%s&apikey=%s",
		f.baseURL, fmpTicker, string(period), f.apiKey)

	body, err := f.doGet(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	var resp []incomeStatementResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: malformed income statement response", financials.ErrNoStatements)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("%w: empty income statement result", financials.ErrNoStatements)
	}

	now := time.Now().UTC()
	stmts := make([]financials.IncomeStatement, len(resp))
	for i, r := range resp {
		year, quarter := parseFMPPeriod(r.CalendarYear, r.Period)
		stmts[i] = financials.IncomeStatement{
			Ticker:            ticker,
			Source:            sourceFMP,
			FiscalYear:        year,
			Quarter:           quarter,
			Period:            period,
			FetchedAt:         now,
			Revenue:           r.Revenue,
			CostOfRevenue:     r.CostOfRevenue,
			GrossProfit:       r.GrossProfit,
			OperatingExpenses: r.OperatingExpenses,
			OperatingIncome:   r.OperatingIncome,
			NetIncome:         r.NetIncome,
			EPS:               r.EPS,
			EPSDiluted:        r.EPSDiluted,
			EBITDA:            r.EBITDA,
			InterestExpense:   r.InterestExpense,
			IncomeTaxExpense:  r.IncomeTaxExpense,
		}
	}
	return stmts, nil
}

// FetchBalanceSheets returns balance sheets for a ticker and period type.
func (f *FMP) FetchBalanceSheets(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	fmpTicker := formatIDXTicker(ticker)
	reqURL := fmt.Sprintf("%s/balance-sheet-statement/%s?period=%s&apikey=%s",
		f.baseURL, fmpTicker, string(period), f.apiKey)

	body, err := f.doGet(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	var resp []balanceSheetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: malformed balance sheet response", financials.ErrNoStatements)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("%w: empty balance sheet result", financials.ErrNoStatements)
	}

	now := time.Now().UTC()
	sheets := make([]financials.BalanceSheet, len(resp))
	for i, r := range resp {
		year, quarter := parseFMPPeriod(r.CalendarYear, r.Period)
		sheets[i] = financials.BalanceSheet{
			Ticker:             ticker,
			Source:             sourceFMP,
			FiscalYear:         year,
			Quarter:            quarter,
			Period:             period,
			FetchedAt:          now,
			TotalAssets:        r.TotalAssets,
			TotalCurrentAssets: r.TotalCurrentAssets,
			CashAndEquivalents: r.CashAndCashEquivalents,
			Receivables:        r.NetReceivables,
			Inventory:          r.Inventory,
			IntangibleAssets:   r.IntangibleAssets,
			TotalLiabilities:   r.TotalLiabilities,
			TotalCurrentLiab:   r.TotalCurrentLiabilities,
			LongTermDebt:       r.LongTermDebt,
			TotalDebt:          r.TotalDebt,
			TotalEquity:        r.TotalStockholdersEquity,
			RetainedEarnings:   r.RetainedEarnings,
			SharesOutstanding:  r.CommonStockSharesOutstanding,
		}
	}
	return sheets, nil
}

// FetchCashFlowStatements returns cash flow statements for a ticker and period type.
func (f *FMP) FetchCashFlowStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	fmpTicker := formatIDXTicker(ticker)
	reqURL := fmt.Sprintf("%s/cash-flow-statement/%s?period=%s&apikey=%s",
		f.baseURL, fmpTicker, string(period), f.apiKey)

	body, err := f.doGet(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	var resp []cashFlowResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: malformed cash flow response", financials.ErrNoStatements)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("%w: empty cash flow result", financials.ErrNoStatements)
	}

	now := time.Now().UTC()
	stmts := make([]financials.CashFlowStatement, len(resp))
	for i, r := range resp {
		year, quarter := parseFMPPeriod(r.CalendarYear, r.Period)
		stmts[i] = financials.CashFlowStatement{
			Ticker:             ticker,
			Source:             sourceFMP,
			FiscalYear:         year,
			Quarter:            quarter,
			Period:             period,
			FetchedAt:          now,
			OperatingCashFlow:  r.OperatingCashFlow,
			CapitalExpenditure: r.CapitalExpenditure,
			FreeCashFlow:       r.FreeCashFlow,
			DividendsPaid:      r.DividendsPaid,
			NetBorrowings:      r.DebtRepayment,
			InvestingCashFlow:  r.NetCashUsedForInvestingActivites,
			FinancingCashFlow:  r.NetCashUsedProvidedByFinancingActivities,
			NetChangeInCash:    r.NetChangeInCash,
		}
	}
	return stmts, nil
}

func (f *FMP) doGet(ctx context.Context, reqURL string) ([]byte, error) {
	if err := f.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := f.client.Do(req) //nolint:gosec // URL is constructed from controlled baseURL + ticker
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if err := mapHTTPError(resp.StatusCode); err != nil {
		return nil, err
	}

	return body, nil
}

func mapHTTPError(code int) error {
	switch {
	case code == http.StatusOK:
		return nil
	case code == http.StatusForbidden:
		return financials.ErrInvalidKey
	case code == http.StatusTooManyRequests:
		return financials.ErrRateLimited
	case code >= 500:
		return financials.ErrSourceDown
	default:
		return fmt.Errorf("unexpected HTTP status %d", code)
	}
}

// formatIDXTicker converts an IDX ticker (e.g. "BBCA") to FMP format ("BBCA.JK").
func formatIDXTicker(ticker string) string {
	if strings.HasSuffix(ticker, ".JK") {
		return ticker
	}
	return ticker + ".JK"
}

// parseFMPPeriod extracts fiscal year and quarter from FMP response fields.
// FMP returns calendarYear as "2024" and period as "FY", "Q1", "Q2", "Q3", "Q4".
func parseFMPPeriod(calendarYear, period string) (int, int) {
	year, _ := strconv.Atoi(calendarYear)

	switch period {
	case "Q1":
		return year, 1
	case "Q2":
		return year, 2
	case "Q3":
		return year, 3
	case "Q4":
		return year, 4
	default:
		return year, 0
	}
}
