package watchlist

import "testing"

func TestNewWatchlist(t *testing.T) {
	w := NewWatchlist("profile-1", "Tech Stocks")

	if w.ID == "" {
		t.Error("expected non-empty ID")
	}
	if w.ProfileID != "profile-1" {
		t.Errorf("ProfileID = %q, want %q", w.ProfileID, "profile-1")
	}
	if w.Name != "Tech Stocks" {
		t.Errorf("Name = %q, want %q", w.Name, "Tech Stocks")
	}
	if w.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if w.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
	if w.CreatedAt != w.UpdatedAt {
		t.Error("expected CreatedAt == UpdatedAt for new watchlist")
	}
}

func TestNewWatchlistGeneratesUniqueIDs(t *testing.T) {
	w1 := NewWatchlist("p1", "A")
	w2 := NewWatchlist("p1", "A")

	if w1.ID == w2.ID {
		t.Error("expected unique IDs for different watchlists")
	}
}

func TestNewItem(t *testing.T) {
	item := NewItem("watchlist-1", "BBCA")

	if item.ID == "" {
		t.Error("expected non-empty ID")
	}
	if item.WatchlistID != "watchlist-1" {
		t.Errorf("WatchlistID = %q, want %q", item.WatchlistID, "watchlist-1")
	}
	if item.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want %q", item.Ticker, "BBCA")
	}
	if item.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNewItemGeneratesUniqueIDs(t *testing.T) {
	i1 := NewItem("w1", "BBCA")
	i2 := NewItem("w1", "BBCA")

	if i1.ID == i2.ID {
		t.Error("expected unique IDs for different items")
	}
}
