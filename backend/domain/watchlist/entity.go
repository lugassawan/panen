package watchlist

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// Watchlist represents a named collection of stock tickers for a profile.
type Watchlist struct {
	ID        string
	ProfileID string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewWatchlist creates a new Watchlist with generated ID and timestamps.
func NewWatchlist(profileID, name string) *Watchlist {
	now := time.Now().UTC()
	return &Watchlist{
		ID:        shared.NewID(),
		ProfileID: profileID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Item represents a single ticker entry within a watchlist.
type Item struct {
	ID          string
	WatchlistID string
	Ticker      string
	CreatedAt   time.Time
}

// NewItem creates a new Item with generated ID and timestamp.
func NewItem(watchlistID, ticker string) *Item {
	return &Item{
		ID:          shared.NewID(),
		WatchlistID: watchlistID,
		Ticker:      ticker,
		CreatedAt:   time.Now().UTC(),
	}
}
