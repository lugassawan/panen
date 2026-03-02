package stock

import "errors"

var (
	ErrInvalidTicker = errors.New("invalid ticker")
	ErrRateLimited   = errors.New("rate limited")
	ErrSourceDown    = errors.New("data source unavailable")
	ErrNoData        = errors.New("no data available")
)
