package stock

import "time"

// Data holds scraped financial data for a single stock ticker.
type Data struct {
	ID            string
	Ticker        string
	Price         float64
	High52Week    float64
	Low52Week     float64
	EPS           float64
	BVPS          float64
	ROE           float64
	DER           float64
	PBV           float64
	PER           float64
	DividendYield float64
	PayoutRatio   float64
	FetchedAt     time.Time
	Source        string
}
