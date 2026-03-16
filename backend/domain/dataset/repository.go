package dataset

import "context"

// Repository defines read access to financial datasets.
type Repository interface {
	FetchFinancials(ctx context.Context, ticker string) (*TickerFinancials, error)
	FetchAllFinancials(ctx context.Context) (FinancialDataset, error)
	FetchSectorMetrics(ctx context.Context, ticker string) (map[string][]SectorMetricEntry, error)
	FetchDefinitions(ctx context.Context) (MetricDefinitions, error)
}
