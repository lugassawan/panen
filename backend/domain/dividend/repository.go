package dividend

import (
	"context"
	"time"
)

// HistoryRepository defines persistence operations for dividend history data.
type HistoryRepository interface {
	BulkUpsert(ctx context.Context, events []DividendEvent) error
	GetByTicker(ctx context.Context, ticker, source string) ([]DividendEvent, error)
	LatestDate(ctx context.Context, ticker, source string) (time.Time, error)
}
