package watchlist

import "errors"

var (
	ErrTickerAlreadyInWatchlist = errors.New("ticker already in watchlist")
	ErrTickerNotInWatchlist     = errors.New("ticker not in watchlist")
)
