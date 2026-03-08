package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/stock"
)

const (
	idxBaseURL     = "https://www.idx.co.id/primary/ListingDev/GetStockInfo"
	idxUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	idxMaxResponse = 2 << 20 // 2 MB
)

// IDXProvider fetches stock data from the Indonesia Stock Exchange (IDX) website.
// It serves as a secondary/fallback provider. Only price data is supported;
// financials and dividend history return ErrNotSupported.
type IDXProvider struct {
	client  *http.Client
	limiter *rate.Limiter
	baseURL string
}

// ErrNotSupported indicates the provider does not support this operation.
var ErrNotSupported = fmt.Errorf("%w: operation not supported by this provider", stock.ErrNoData)

// IDXOption configures the IDXProvider.
type IDXOption func(*IDXProvider)

// WithIDXBaseURL overrides the base URL (useful for tests).
func WithIDXBaseURL(u string) IDXOption {
	return func(p *IDXProvider) { p.baseURL = u }
}

// WithIDXRateLimit sets custom rate limiter parameters.
func WithIDXRateLimit(rps float64, burst int) IDXOption {
	return func(p *IDXProvider) { p.limiter = rate.NewLimiter(rate.Limit(rps), burst) }
}

// NewIDXProvider creates a new IDX data provider.
func NewIDXProvider(opts ...IDXOption) *IDXProvider {
	p := &IDXProvider{
		client:  &http.Client{Timeout: 15 * time.Second},
		limiter: rate.NewLimiter(1, 3),
		baseURL: idxBaseURL,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Source returns the provider identifier.
func (p *IDXProvider) Source() string { return "idx" }

// idxStockResponse maps the relevant fields from the IDX stock info API.
type idxStockResponse struct {
	LastPrice  float64 `json:"LastPrice"`
	High52Week float64 `json:"High52Week"`
	Low52Week  float64 `json:"Low52Week"`
	PER        float64 `json:"PER"`
	PBV        float64 `json:"PBV"`
}

// FetchPrice returns the current price and 52-week range from IDX.
func (p *IDXProvider) FetchPrice(ctx context.Context, ticker string) (*stock.PriceResult, error) {
	data, err := p.fetchStockInfo(ctx, ticker)
	if err != nil {
		return nil, err
	}

	if data.LastPrice == 0 {
		return nil, fmt.Errorf("%w: missing price from IDX", stock.ErrNoData)
	}

	return &stock.PriceResult{
		Price:      data.LastPrice,
		High52Week: data.High52Week,
		Low52Week:  data.Low52Week,
	}, nil
}

// FetchFinancials returns partial financial metrics from IDX.
// Only PER and PBV are available; other fields are zero.
func (p *IDXProvider) FetchFinancials(ctx context.Context, ticker string) (*stock.FinancialResult, error) {
	data, err := p.fetchStockInfo(ctx, ticker)
	if err != nil {
		return nil, err
	}

	return &stock.FinancialResult{
		PER: data.PER,
		PBV: data.PBV,
	}, nil
}

// FetchPriceHistory is not supported by the IDX provider.
func (p *IDXProvider) FetchPriceHistory(_ context.Context, _ string) ([]stock.PricePoint, error) {
	return nil, ErrNotSupported
}

// FetchDividendHistory is not supported by the IDX provider.
func (p *IDXProvider) FetchDividendHistory(_ context.Context, _ string) ([]dividend.DividendEvent, error) {
	return nil, ErrNotSupported
}

func (p *IDXProvider) fetchStockInfo(ctx context.Context, ticker string) (*idxStockResponse, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	// Remove .JK suffix if present (IDX uses raw tickers).
	ticker = strings.TrimSuffix(ticker, ".JK")
	// Index tickers are not supported.
	if strings.HasPrefix(ticker, "^") {
		return nil, fmt.Errorf("%w: index tickers not supported by IDX provider", stock.ErrNoData)
	}

	reqURL := fmt.Sprintf("%s?code=%s", p.baseURL, ticker)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating IDX request: %w", err)
	}
	req.Header.Set("User-Agent", idxUserAgent)

	resp, err := p.client.Do(req) //nolint:gosec // URL is constructed from controlled baseURL + ticker query param
	if err != nil {
		return nil, fmt.Errorf("fetching from IDX: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, stock.ErrInvalidTicker
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, stock.ErrRateLimited
	}
	if resp.StatusCode >= 500 {
		return nil, stock.ErrSourceDown
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IDX returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, idxMaxResponse))
	if err != nil {
		return nil, fmt.Errorf("reading IDX response: %w", err)
	}

	var data idxStockResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("%w: malformed IDX response", stock.ErrNoData)
	}

	return &data, nil
}
