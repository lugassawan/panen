package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
	"github.com/lugassawan/panen/backend/domain/watchlist"
)

type watchlistTestFixture struct {
	repo     *WatchlistRepo
	itemRepo *WatchlistItemRepo
	profile  *user.Profile
	ctx      context.Context
	now      time.Time
}

func setupWatchlistTest(t *testing.T) watchlistTestFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	userRepo := NewUserRepo(db)

	p := &user.Profile{
		ID:        shared.NewID(),
		Name:      "User",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := userRepo.Create(ctx, p); err != nil {
		t.Fatalf("create profile: %v", err)
	}

	return watchlistTestFixture{
		repo:     NewWatchlistRepo(db),
		itemRepo: NewWatchlistItemRepo(db),
		profile:  p,
		ctx:      ctx,
		now:      now,
	}
}

func TestWatchlistRepoCreateAndGetByID(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Tech Stocks",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Tech Stocks" {
		t.Errorf("Name = %q, want %q", got.Name, "Tech Stocks")
	}
	if got.ProfileID != f.profile.ID {
		t.Errorf("ProfileID = %q, want %q", got.ProfileID, f.profile.ID)
	}
}

func TestWatchlistRepoGetByIDNotFound(t *testing.T) {
	f := setupWatchlistTest(t)

	_, err := f.repo.GetByID(f.ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestWatchlistRepoDuplicateNameConstraint(t *testing.T) {
	f := setupWatchlistTest(t)

	w1 := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Favorites",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w1); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	w2 := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Favorites",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w2); err == nil {
		t.Fatal("Create() duplicate name for same profile should fail")
	}
}

func TestWatchlistRepoListByProfileID(t *testing.T) {
	f := setupWatchlistTest(t)

	w1 := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Alpha",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	w2 := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Beta",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w1); err != nil {
		t.Fatalf("Create() w1 error = %v", err)
	}
	if err := f.repo.Create(f.ctx, w2); err != nil {
		t.Fatalf("Create() w2 error = %v", err)
	}

	list, err := f.repo.ListByProfileID(f.ctx, f.profile.ID)
	if err != nil {
		t.Fatalf("ListByProfileID() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListByProfileID() count = %d, want 2", len(list))
	}
}

func TestWatchlistRepoUpdate(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Old Name",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updatedAt := f.now.Add(time.Hour)
	w.Name = "New Name"
	w.UpdatedAt = updatedAt
	if err := f.repo.Update(f.ctx, w); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := f.repo.GetByID(f.ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "New Name" {
		t.Errorf("Name = %q, want %q", got.Name, "New Name")
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, updatedAt)
	}
	if got.ProfileID != f.profile.ID {
		t.Errorf("ProfileID changed unexpectedly: got %q", got.ProfileID)
	}
}

func TestWatchlistRepoDelete(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "To Delete",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := f.repo.Delete(f.ctx, w.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := f.repo.GetByID(f.ctx, w.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after delete error = %v, want ErrNotFound", err)
	}
}

func TestWatchlistRepoCascadeDeleteItems(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Cascade Test",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() watchlist error = %v", err)
	}

	tickers := []string{"BBCA", "BBRI", "BMRI"}
	for _, ticker := range tickers {
		item := &watchlist.Item{
			ID:          shared.NewID(),
			WatchlistID: w.ID,
			Ticker:      ticker,
			CreatedAt:   f.now,
		}
		if err := f.itemRepo.Add(f.ctx, item); err != nil {
			t.Fatalf("Add() item %s error = %v", ticker, err)
		}
	}

	before, err := f.itemRepo.ListByWatchlistID(f.ctx, w.ID)
	if err != nil {
		t.Fatalf("ListByWatchlistID() before delete error = %v", err)
	}
	if len(before) != len(tickers) {
		t.Fatalf("expected %d items before delete, got %d", len(tickers), len(before))
	}

	if err := f.repo.Delete(f.ctx, w.ID); err != nil {
		t.Fatalf("Delete() watchlist error = %v", err)
	}

	_, err = f.repo.GetByID(f.ctx, w.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() watchlist after delete error = %v, want ErrNotFound", err)
	}

	after, err := f.itemRepo.ListByWatchlistID(f.ctx, w.ID)
	if err != nil {
		t.Fatalf("ListByWatchlistID() after cascade delete error = %v", err)
	}
	if len(after) != 0 {
		t.Errorf("expected 0 items after cascade delete, got %d", len(after))
	}
}

func TestWatchlistItemRepoAddAndList(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "My List",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() watchlist error = %v", err)
	}

	item := &watchlist.Item{
		ID:          shared.NewID(),
		WatchlistID: w.ID,
		Ticker:      "TLKM",
		CreatedAt:   f.now,
	}
	if err := f.itemRepo.Add(f.ctx, item); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	items, err := f.itemRepo.ListByWatchlistID(f.ctx, w.ID)
	if err != nil {
		t.Fatalf("ListByWatchlistID() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListByWatchlistID() count = %d, want 1", len(items))
	}
	if items[0].Ticker != "TLKM" {
		t.Errorf("Ticker = %q, want %q", items[0].Ticker, "TLKM")
	}
}

func TestWatchlistItemRepoDuplicateTickerConstraint(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Dup Ticker List",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() watchlist error = %v", err)
	}

	item1 := &watchlist.Item{
		ID:          shared.NewID(),
		WatchlistID: w.ID,
		Ticker:      "BBRI",
		CreatedAt:   f.now,
	}
	if err := f.itemRepo.Add(f.ctx, item1); err != nil {
		t.Fatalf("Add() item1 error = %v", err)
	}

	item2 := &watchlist.Item{
		ID:          shared.NewID(),
		WatchlistID: w.ID,
		Ticker:      "BBRI",
		CreatedAt:   f.now,
	}
	if err := f.itemRepo.Add(f.ctx, item2); err == nil {
		t.Fatal("Add() duplicate ticker should fail")
	}
}

func TestWatchlistItemRepoRemove(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Remove Test",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() watchlist error = %v", err)
	}

	item1 := &watchlist.Item{
		ID:          shared.NewID(),
		WatchlistID: w.ID,
		Ticker:      "BMRI",
		CreatedAt:   f.now,
	}
	item2 := &watchlist.Item{
		ID:          shared.NewID(),
		WatchlistID: w.ID,
		Ticker:      "TLKM",
		CreatedAt:   f.now,
	}
	if err := f.itemRepo.Add(f.ctx, item1); err != nil {
		t.Fatalf("Add() item1 error = %v", err)
	}
	if err := f.itemRepo.Add(f.ctx, item2); err != nil {
		t.Fatalf("Add() item2 error = %v", err)
	}

	if err := f.itemRepo.Remove(f.ctx, w.ID, item1.Ticker); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	items, err := f.itemRepo.ListByWatchlistID(f.ctx, w.ID)
	if err != nil {
		t.Fatalf("ListByWatchlistID() error = %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item after remove, got %d", len(items))
	}
	if items[0].Ticker != "TLKM" {
		t.Errorf("remaining ticker = %q, want %q", items[0].Ticker, "TLKM")
	}

	wl, err := f.repo.GetByID(f.ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID() watchlist after item remove error = %v", err)
	}
	if wl.ID != w.ID {
		t.Errorf("watchlist still exists with ID = %q, want %q", wl.ID, w.ID)
	}
}

func TestWatchlistItemRepoRemoveNotFound(t *testing.T) {
	f := setupWatchlistTest(t)

	// Create a watchlist to remove from — ticker doesn't exist
	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Remove Not Found",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() watchlist error = %v", err)
	}

	err := f.itemRepo.Remove(f.ctx, w.ID, "NONEXIST")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Remove() error = %v, want ErrNotFound", err)
	}
}

func TestWatchlistItemRepoExistsByWatchlistAndTicker(t *testing.T) {
	f := setupWatchlistTest(t)

	w := &watchlist.Watchlist{
		ID:        shared.NewID(),
		ProfileID: f.profile.ID,
		Name:      "Exists Test",
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	if err := f.repo.Create(f.ctx, w); err != nil {
		t.Fatalf("Create() watchlist error = %v", err)
	}

	exists, err := f.itemRepo.ExistsByWatchlistAndTicker(f.ctx, w.ID, "ASII")
	if err != nil {
		t.Fatalf("ExistsByWatchlistAndTicker() error = %v", err)
	}
	if exists {
		t.Error("ExistsByWatchlistAndTicker() = true, want false before adding")
	}

	item := &watchlist.Item{
		ID:          shared.NewID(),
		WatchlistID: w.ID,
		Ticker:      "ASII",
		CreatedAt:   f.now,
	}
	if err := f.itemRepo.Add(f.ctx, item); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	exists, err = f.itemRepo.ExistsByWatchlistAndTicker(f.ctx, w.ID, "ASII")
	if err != nil {
		t.Fatalf("ExistsByWatchlistAndTicker() error = %v", err)
	}
	if !exists {
		t.Error("ExistsByWatchlistAndTicker() = false, want true after adding")
	}
}
