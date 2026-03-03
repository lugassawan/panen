package watchlist

import "context"

// Repository defines persistence operations for watchlists.
type Repository interface {
	Create(ctx context.Context, w *Watchlist) error
	GetByID(ctx context.Context, id string) (*Watchlist, error)
	ListByProfileID(ctx context.Context, profileID string) ([]*Watchlist, error)
	Update(ctx context.Context, w *Watchlist) error
	Delete(ctx context.Context, id string) error
}

// ItemRepository defines persistence operations for watchlist items.
type ItemRepository interface {
	Add(ctx context.Context, item *Item) error
	Remove(ctx context.Context, watchlistID, ticker string) error
	ListByWatchlistID(ctx context.Context, watchlistID string) ([]*Item, error)
	ExistsByWatchlistAndTicker(ctx context.Context, watchlistID, ticker string) (bool, error)
}
