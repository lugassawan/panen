package database

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/dividend"
)

func newDividendHistoryRepo(t *testing.T) (*DividendHistoryRepo, context.Context) {
	t.Helper()
	db := newTestDB(t)
	return NewDividendHistoryRepo(db), context.Background()
}

func TestDividendHistoryRepoBulkUpsertAndGet(t *testing.T) {
	repo, ctx := newDividendHistoryRepo(t)

	events := []dividend.DividendEvent{
		{Ticker: "BBCA", ExDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50, Source: "yahoo"},
		{Ticker: "BBCA", ExDate: time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 60, Source: "yahoo"},
	}

	if err := repo.BulkUpsert(ctx, events); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].Amount != 50 {
		t.Errorf("got[0].Amount = %v, want 50", got[0].Amount)
	}
	if got[1].Amount != 60 {
		t.Errorf("got[1].Amount = %v, want 60", got[1].Amount)
	}
}

func TestDividendHistoryRepoBulkUpsertUpdatesExisting(t *testing.T) {
	repo, ctx := newDividendHistoryRepo(t)

	events := []dividend.DividendEvent{
		{Ticker: "BBCA", ExDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50, Source: "yahoo"},
	}
	if err := repo.BulkUpsert(ctx, events); err != nil {
		t.Fatalf("BulkUpsert() insert error = %v", err)
	}

	events[0].Amount = 55
	events[0].ID = ""
	if err := repo.BulkUpsert(ctx, events); err != nil {
		t.Fatalf("BulkUpsert() update error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].Amount != 55 {
		t.Errorf("Amount = %v, want 55", got[0].Amount)
	}
}

func TestDividendHistoryRepoLatestDate(t *testing.T) {
	repo, ctx := newDividendHistoryRepo(t)

	t.Run("returns zero when empty", func(t *testing.T) {
		latest, err := repo.LatestDate(ctx, "BBCA", "yahoo")
		if err != nil {
			t.Fatalf("LatestDate() error = %v", err)
		}
		if !latest.IsZero() {
			t.Errorf("LatestDate() = %v, want zero", latest)
		}
	})

	t.Run("returns max date after insert", func(t *testing.T) {
		events := []dividend.DividendEvent{
			{Ticker: "TLKM", ExDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50, Source: "yahoo"},
			{Ticker: "TLKM", ExDate: time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), Amount: 55, Source: "yahoo"},
			{Ticker: "TLKM", ExDate: time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 60, Source: "yahoo"},
		}
		if err := repo.BulkUpsert(ctx, events); err != nil {
			t.Fatalf("BulkUpsert() error = %v", err)
		}
		latest, err := repo.LatestDate(ctx, "TLKM", "yahoo")
		if err != nil {
			t.Fatalf("LatestDate() error = %v", err)
		}
		want := time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC)
		if !latest.Equal(want) {
			t.Errorf("LatestDate() = %v, want %v", latest, want)
		}
	})
}

func TestDividendHistoryRepoGetByTickerEmpty(t *testing.T) {
	repo, ctx := newDividendHistoryRepo(t)

	got, err := repo.GetByTicker(ctx, "NONEXIST", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}
