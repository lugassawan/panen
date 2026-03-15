package sectorsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/lugassawan/panen/backend/domain/financials"
)

const (
	defaultBaseURL  = "https://api.sectors.app/v1"
	maxResponseSize = 5 << 20 // 5 MB
	sourceSectors   = "sectors"
)

// SectorsApp fetches financial statement data from the Sectors.app API.
type SectorsApp struct {
	client  *http.Client
	limiter *rate.Limiter
	baseURL string
	apiKey  string
}

// Option configures the SectorsApp provider.
type Option func(*SectorsApp)

// WithBaseURL overrides the base URL (useful for tests).
func WithBaseURL(u string) Option {
	return func(s *SectorsApp) { s.baseURL = u }
}

// WithRateLimit sets custom rate limiter parameters.
func WithRateLimit(rps float64, burst int) Option {
	return func(s *SectorsApp) { s.limiter = rate.NewLimiter(rate.Limit(rps), burst) }
}

// WithHTTPClient overrides the HTTP client (useful for tests).
func WithHTTPClient(c *http.Client) Option {
	return func(s *SectorsApp) { s.client = c }
}

// NewSectorsApp creates a SectorsApp provider with the given API key and options.
func NewSectorsApp(apiKey string, opts ...Option) *SectorsApp {
	s := &SectorsApp{
		limiter: rate.NewLimiter(2, 5),
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
	}
	for _, o := range opts {
		o(s)
	}
	if s.client == nil {
		s.client = &http.Client{}
	}
	return s
}

// Source returns the provider identifier.
func (s *SectorsApp) Source() string { return sourceSectors }

// FetchIncomeStatements returns income statements for a ticker and period type.
func (s *SectorsApp) FetchIncomeStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	reports, err := s.fetchReports(ctx, ticker, period)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	stmts := make([]financials.IncomeStatement, len(reports))
	for i, r := range reports {
		stmts[i] = financials.IncomeStatement{
			Ticker:            ticker,
			Source:            sourceSectors,
			FiscalYear:        r.Year,
			Quarter:           r.Quarter,
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
func (s *SectorsApp) FetchBalanceSheets(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	reports, err := s.fetchReports(ctx, ticker, period)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	sheets := make([]financials.BalanceSheet, len(reports))
	for i, r := range reports {
		sheets[i] = financials.BalanceSheet{
			Ticker:             ticker,
			Source:             sourceSectors,
			FiscalYear:         r.Year,
			Quarter:            r.Quarter,
			Period:             period,
			FetchedAt:          now,
			TotalAssets:        r.TotalAssets,
			TotalCurrentAssets: r.TotalCurrentAssets,
			CashAndEquivalents: r.CashAndEquivalents,
			Receivables:        r.Receivables,
			Inventory:          r.Inventory,
			IntangibleAssets:   r.IntangibleAssets,
			TotalLiabilities:   r.TotalLiabilities,
			TotalCurrentLiab:   r.TotalCurrentLiab,
			LongTermDebt:       r.LongTermDebt,
			TotalDebt:          r.TotalDebt,
			TotalEquity:        r.TotalEquity,
			RetainedEarnings:   r.RetainedEarnings,
			SharesOutstanding:  r.SharesOutstanding,
		}
	}
	return sheets, nil
}

// FetchCashFlowStatements returns cash flow statements for a ticker and period type.
func (s *SectorsApp) FetchCashFlowStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	reports, err := s.fetchReports(ctx, ticker, period)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	stmts := make([]financials.CashFlowStatement, len(reports))
	for i, r := range reports {
		stmts[i] = financials.CashFlowStatement{
			Ticker:             ticker,
			Source:             sourceSectors,
			FiscalYear:         r.Year,
			Quarter:            r.Quarter,
			Period:             period,
			FetchedAt:          now,
			OperatingCashFlow:  r.OperatingCashFlow,
			CapitalExpenditure: r.CapitalExpenditure,
			FreeCashFlow:       r.FreeCashFlow,
			DividendsPaid:      r.DividendsPaid,
			NetBorrowings:      r.NetBorrowings,
			InvestingCashFlow:  r.InvestingCashFlow,
			FinancingCashFlow:  r.FinancingCashFlow,
			NetChangeInCash:    r.NetChangeInCash,
		}
	}
	return stmts, nil
}

func (s *SectorsApp) fetchReports(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financialReportResponse, error) {
	reqURL := fmt.Sprintf("%s/company/financials/%s/?period=%s", s.baseURL, ticker, string(period))

	body, err := s.doGet(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	var reports []financialReportResponse
	if err := json.Unmarshal(body, &reports); err != nil {
		return nil, fmt.Errorf("%w: malformed financial report response", financials.ErrNoStatements)
	}

	if len(reports) == 0 {
		return nil, fmt.Errorf("%w: empty financial report result", financials.ErrNoStatements)
	}

	return reports, nil
}

func (s *SectorsApp) doGet(ctx context.Context, reqURL string) ([]byte, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req) //nolint:gosec // URL is constructed from controlled baseURL + ticker
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
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		return financials.ErrInvalidKey
	case code == http.StatusTooManyRequests:
		return financials.ErrRateLimited
	case code >= 500:
		return financials.ErrSourceDown
	default:
		return fmt.Errorf("unexpected HTTP status %d", code)
	}
}
