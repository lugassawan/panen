package stock

import "context"

// Repository defines persistence operations for scraped stock data.
type Repository interface {
	Upsert(ctx context.Context, data *Data) error
	GetByTicker(ctx context.Context, ticker string) (*Data, error)
	GetByTickerAndSource(ctx context.Context, ticker string, source string) (*Data, error)
	DeleteOlderThan(ctx context.Context, before string) (int64, error)
}
