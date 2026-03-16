package dataset

// TickerFinancials holds financial time series for a single stock.
type TickerFinancials struct {
	Revenue         []TimeSeriesEntry
	NetIncome       []TimeSeriesEntry
	TotalAssets     []TimeSeriesEntry
	TotalEquity     []TimeSeriesEntry
	TotalDebt       []TimeSeriesEntry
	OperatingIncome []TimeSeriesEntry
}

// TimeSeriesEntry is a single data point in a financial time series.
type TimeSeriesEntry struct {
	Year    int
	Quarter int
	Value   float64
}

// SectorMetricEntry is a single sector-specific metric value.
type SectorMetricEntry struct {
	Year    int
	Quarter int
	Value   float64
	Source  string // "auto" or "manual"
}

// MetricDefinitions maps sector names to their metric lists.
type MetricDefinitions map[string][]string
