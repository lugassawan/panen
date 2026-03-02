package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"

	"golang.org/x/time/rate"

	"github.com/lugassawan/panen/backend/domain/stock"
)

const (
	defaultBaseURL = "https://query1.finance.yahoo.com"
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	maxResponseSize = 5 << 20 // 5 MB
)

// Yahoo fetches stock data from Yahoo Finance JSON APIs.
type Yahoo struct {
	client  *http.Client
	limiter *rate.Limiter
	baseURL string
}

// Option configures the Yahoo provider.
type Option func(*Yahoo)

// WithClient sets a custom HTTP client.
func WithClient(c *http.Client) Option {
	return func(y *Yahoo) { y.client = c }
}

// WithBaseURL overrides the base URL (useful for tests).
func WithBaseURL(url string) Option {
	return func(y *Yahoo) { y.baseURL = url }
}

// WithRateLimit sets custom rate limiter parameters.
func WithRateLimit(rps float64, burst int) Option {
	return func(y *Yahoo) { y.limiter = rate.NewLimiter(rate.Limit(rps), burst) }
}

// NewYahoo creates a Yahoo provider with the given options.
func NewYahoo(opts ...Option) *Yahoo {
	y := &Yahoo{
		client:  http.DefaultClient,
		limiter: rate.NewLimiter(2, 5),
		baseURL: defaultBaseURL,
	}
	for _, o := range opts {
		o(y)
	}
	return y
}

// Source returns the provider identifier.
func (y *Yahoo) Source() string { return "yahoo" }

// FetchPrice returns the current price and 52-week range for a ticker.
func (y *Yahoo) FetchPrice(ctx context.Context, ticker string) (*stock.PriceResult, error) {
	encoded := url.PathEscape(FormatIDX(ticker))
	reqURL := fmt.Sprintf("%s/v8/finance/chart/%s?range=1y&interval=1d", y.baseURL, encoded)

	body, err := y.doGet(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	var resp chartResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: malformed chart response", stock.ErrNoData)
	}

	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("%w: %s", stock.ErrNoData, resp.Chart.Error.Description)
	}

	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("%w: empty chart result", stock.ErrNoData)
	}

	result := resp.Chart.Result[0]
	price := result.Meta.RegularMarketPrice
	if price == 0 {
		return nil, fmt.Errorf("%w: missing price", stock.ErrNoData)
	}

	high52, low52 := compute52WeekRange(result.Indicators)

	return &stock.PriceResult{
		Price:      price,
		High52Week: high52,
		Low52Week:  low52,
	}, nil
}

// FetchFinancials returns fundamental financial metrics for a ticker.
func (y *Yahoo) FetchFinancials(ctx context.Context, ticker string) (*stock.FinancialResult, error) {
	encoded := url.PathEscape(FormatIDX(ticker))
	reqURL := fmt.Sprintf(
		"%s/v10/finance/quoteSummary/%s?modules=defaultKeyStatistics,financialData,summaryDetail",
		y.baseURL, encoded,
	)

	body, err := y.doGet(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	var resp quoteSummaryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: malformed quoteSummary response", stock.ErrNoData)
	}

	if resp.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("%w: %s", stock.ErrNoData, resp.QuoteSummary.Error.Description)
	}

	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("%w: empty quoteSummary result", stock.ErrNoData)
	}

	r := resp.QuoteSummary.Result[0]

	return &stock.FinancialResult{
		EPS:           r.DefaultKeyStatistics.TrailingEps.Raw,
		BVPS:          r.DefaultKeyStatistics.BookValue.Raw,
		ROE:           r.FinancialData.ReturnOnEquity.Raw * 100,
		DER:           r.FinancialData.DebtToEquity.Raw,
		PBV:           r.DefaultKeyStatistics.PriceToBook.Raw,
		PER:           r.DefaultKeyStatistics.TrailingPE.Raw,
		DividendYield: r.SummaryDetail.DividendYield.Raw * 100,
		PayoutRatio:   r.SummaryDetail.PayoutRatio.Raw * 100,
	}, nil
}

func (y *Yahoo) doGet(ctx context.Context, url string) ([]byte, error) {
	if err := y.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := y.client.Do(req) //nolint:gosec // URL is constructed from controlled baseURL + ticker path segment
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, mapHTTPError(resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return body, nil
}

func mapHTTPError(code int) error {
	switch {
	case code == http.StatusNotFound:
		return stock.ErrInvalidTicker
	case code == http.StatusTooManyRequests:
		return stock.ErrRateLimited
	case code >= 500:
		return stock.ErrSourceDown
	default:
		return fmt.Errorf("unexpected HTTP status %d", code)
	}
}

func compute52WeekRange(ind indicators) (high, low float64) {
	if len(ind.Quote) == 0 {
		return 0, 0
	}

	q := ind.Quote[0]
	high = 0
	low = math.MaxFloat64

	for _, v := range q.High {
		if v != nil && *v > high {
			high = *v
		}
	}

	for _, v := range q.Low {
		if v != nil && *v < low {
			low = *v
		}
	}

	if low == math.MaxFloat64 {
		low = 0
	}

	return high, low
}
