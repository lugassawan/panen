package valuation

// BandStats holds descriptive statistics for a set of historical values.
type BandStats struct {
	Min, Max, Avg, Median float64
}

// ValuationInput contains all data needed to evaluate a stock.
type ValuationInput struct {
	Ticker      string
	Price       float64     // current market price
	EPS         float64     // trailing EPS
	BVPS        float64     // book value per share
	PBV         float64     // current price-to-book
	PER         float64     // current price-to-earnings
	RiskProfile RiskProfile // risk tolerance for margin of safety
	HistPBV     []float64   // historical PBV values (for band analysis)
	HistPER     []float64   // historical PER values (for band analysis)
}

// ValuationResult holds the computed valuation metrics for a stock.
type ValuationResult struct {
	Ticker         string
	GrahamNumber   float64    // √(22.5 × EPS × BVPS); 0 if not applicable
	PBVBand        *BandStats // nil if insufficient data
	PERBand        *BandStats // nil if insufficient data
	MarginOfSafety float64    // percentage (e.g. 50.0 for 50%)
	EntryPrice     float64    // intrinsic value adjusted down by margin
	ExitTarget     float64    // upper band target price
	Verdict        Verdict
}
