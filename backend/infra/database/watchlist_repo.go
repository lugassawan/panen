package database

import (
	"context"
	"database/sql"

	"github.com/lugassawan/panen/backend/domain/watchlist"
)

const (
	watchlistInsert = `INSERT INTO watchlists
		(id, profile_id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`
	watchlistGetByID = `SELECT id, profile_id, name, created_at, updated_at
		FROM watchlists WHERE id = ?`
	watchlistListByProfileID = `SELECT id, profile_id, name, created_at, updated_at
		FROM watchlists WHERE profile_id = ? ORDER BY name`
	watchlistUpdate = `UPDATE watchlists SET name = ?, updated_at = ?
		WHERE id = ?`
	watchlistDelete  = `DELETE FROM watchlists WHERE id = ?`
	watchlistItemAdd = `INSERT INTO watchlist_items
		(id, watchlist_id, ticker, created_at)
		VALUES (?, ?, ?, ?)`
	watchlistItemRemove            = `DELETE FROM watchlist_items WHERE watchlist_id = ? AND ticker = ?`
	watchlistItemListByWatchlistID = `SELECT id, watchlist_id, ticker, created_at
		FROM watchlist_items WHERE watchlist_id = ? ORDER BY ticker`
	watchlistItemExists = `SELECT COUNT(*) FROM watchlist_items
		WHERE watchlist_id = ? AND ticker = ?`
)

// WatchlistRepo implements watchlist.Repository.
type WatchlistRepo struct {
	db *sql.DB
}

// NewWatchlistRepo creates a new WatchlistRepo.
func NewWatchlistRepo(db *sql.DB) *WatchlistRepo {
	return &WatchlistRepo{db: db}
}

func (r *WatchlistRepo) Create(ctx context.Context, w *watchlist.Watchlist) error {
	_, err := r.db.ExecContext(ctx, watchlistInsert,
		w.ID, w.ProfileID, w.Name,
		formatTime(w.CreatedAt), formatTime(w.UpdatedAt))
	return err
}

func (r *WatchlistRepo) GetByID(ctx context.Context, id string) (*watchlist.Watchlist, error) {
	return QueryRow(ctx, r.db, watchlistGetByID, scanWatchlist, id)
}

func (r *WatchlistRepo) ListByProfileID(ctx context.Context, profileID string) ([]*watchlist.Watchlist, error) {
	return QueryAll(ctx, r.db, watchlistListByProfileID, scanWatchlist, profileID)
}

func (r *WatchlistRepo) Update(ctx context.Context, w *watchlist.Watchlist) error {
	res, err := r.db.ExecContext(ctx, watchlistUpdate,
		w.Name, formatTime(w.UpdatedAt), w.ID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *WatchlistRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, watchlistDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

// WatchlistItemRepo implements watchlist.ItemRepository.
type WatchlistItemRepo struct {
	db *sql.DB
}

// NewWatchlistItemRepo creates a new WatchlistItemRepo.
func NewWatchlistItemRepo(db *sql.DB) *WatchlistItemRepo {
	return &WatchlistItemRepo{db: db}
}

func (r *WatchlistItemRepo) Add(ctx context.Context, item *watchlist.Item) error {
	_, err := r.db.ExecContext(ctx, watchlistItemAdd,
		item.ID, item.WatchlistID, item.Ticker, formatTime(item.CreatedAt))
	return err
}

func (r *WatchlistItemRepo) Remove(ctx context.Context, watchlistID, ticker string) error {
	res, err := r.db.ExecContext(ctx, watchlistItemRemove, watchlistID, ticker)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *WatchlistItemRepo) ListByWatchlistID(ctx context.Context, watchlistID string) ([]*watchlist.Item, error) {
	return QueryAll(ctx, r.db, watchlistItemListByWatchlistID, scanWatchlistItem, watchlistID)
}

func (r *WatchlistItemRepo) ExistsByWatchlistAndTicker(ctx context.Context, watchlistID, ticker string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, watchlistItemExists, watchlistID, ticker).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func scanWatchlist(scan func(dest ...any) error) (*watchlist.Watchlist, error) {
	var w watchlist.Watchlist
	var createdAt, updatedAt string
	if err := scan(&w.ID, &w.ProfileID, &w.Name, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	var err error
	if w.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if w.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &w, nil
}

func scanWatchlistItem(scan func(dest ...any) error) (*watchlist.Item, error) {
	var item watchlist.Item
	var createdAt string
	if err := scan(&item.ID, &item.WatchlistID, &item.Ticker, &createdAt); err != nil {
		return nil, err
	}
	var err error
	if item.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	return &item, nil
}
