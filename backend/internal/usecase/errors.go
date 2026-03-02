package usecase

import "errors"

var (
	ErrEmptyTicker  = errors.New("ticker is required")
	ErrEmptyName    = errors.New("name is required")
	ErrInvalidPrice = errors.New("price must be positive")
	ErrInvalidLots  = errors.New("lots must be positive")
	ErrInvalidFee   = errors.New("fee must not be negative")
	ErrInvalidMode  = errors.New("invalid portfolio mode")
	ErrInvalidRisk  = errors.New("invalid risk profile")
	ErrNoStockData  = errors.New("no cached stock data available")
)
