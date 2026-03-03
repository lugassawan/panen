package usecase

import "errors"

var (
	ErrEmptyID       = errors.New("id is required")
	ErrEmptyTicker   = errors.New("ticker is required")
	ErrEmptyName     = errors.New("name is required")
	ErrInvalidPrice  = errors.New("price must be positive")
	ErrInvalidLots   = errors.New("lots must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNoStockData   = errors.New("no cached stock data available")
	ErrHasDependents = errors.New("has dependent portfolios")
)
