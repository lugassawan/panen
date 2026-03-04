package checklist

import "errors"

// Sentinel errors for checklist operations.
var (
	ErrChecklistNotReady = errors.New("checklist not ready")
	ErrNoStockData       = errors.New("no stock data available")
	ErrNoHolding         = errors.New("no holding found")
	ErrNoValuation       = errors.New("no valuation available")
)
