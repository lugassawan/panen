package trailingstop

import "time"

// HoldingPeak tracks the highest observed price for a holding.
type HoldingPeak struct {
	ID        string
	HoldingID string
	PeakPrice float64
	UpdatedAt time.Time
}

// TrailingStopResult holds the computed trailing stop data for a single holding.
type TrailingStopResult struct {
	PeakPrice        float64
	StopPct          float64
	StopPrice        float64
	Triggered        bool
	FundamentalExits []FundamentalExit
}

// FundamentalExit describes a fundamental deterioration criterion.
type FundamentalExit struct {
	Key       string
	Label     string
	Detail    string
	Triggered bool
}
