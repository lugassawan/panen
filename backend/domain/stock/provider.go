package stock

import "context"

// PriceResult holds price data fetched from an external source.
type PriceResult struct {
	Price      float64
	High52Week float64
	Low52Week  float64
}

// FinancialResult holds fundamental financial metrics.
type FinancialResult struct {
	EPS           float64
	BVPS          float64
	ROE           float64
	DER           float64
	PBV           float64
	PER           float64
	DividendYield float64
	PayoutRatio   float64
}

// DataProvider defines operations for fetching stock data from external sources.
type DataProvider interface {
	// Source returns the provider identifier (e.g. "yahoo").
	Source() string
	// FetchPrice returns current price and 52-week range for a ticker.
	FetchPrice(ctx context.Context, ticker string) (*PriceResult, error)
	// FetchFinancials returns fundamental financial metrics for a ticker.
	FetchFinancials(ctx context.Context, ticker string) (*FinancialResult, error)
}
