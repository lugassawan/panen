package financials

import "errors"

var (
	ErrNoStatements = errors.New("no financial statements available")
	ErrRateLimited  = errors.New("rate limited")
	ErrSourceDown   = errors.New("financial data source unavailable")
	ErrInvalidKey   = errors.New("invalid API key")
)
