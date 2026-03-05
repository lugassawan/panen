package stock

import (
	"context"
	"time"
)

// PricePoint represents a single day's OHLCV data for a stock.
type PricePoint struct {
	ID     string
	Ticker string
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
	Source string
}

// PriceHistoryRepository defines persistence operations for price history data.
type PriceHistoryRepository interface {
	BulkUpsert(ctx context.Context, points []PricePoint) error
	GetByTicker(ctx context.Context, ticker, source string) ([]PricePoint, error)
	LatestDate(ctx context.Context, ticker, source string) (time.Time, error)
	DeleteByTicker(ctx context.Context, ticker string) error
}
