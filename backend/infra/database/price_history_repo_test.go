package database

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

func newPriceHistoryRepo(t *testing.T) (*PriceHistoryRepo, context.Context) {
	t.Helper()
	db := newTestDB(t)
	return NewPriceHistoryRepo(db), context.Background()
}

func TestPriceHistoryRepoBulkUpsertAndGet(t *testing.T) {
	repo, ctx := newPriceHistoryRepo(t)

	points := []stock.PricePoint{
		{
			Ticker: "BBCA",
			Date:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			Open:   9000,
			High:   9200,
			Low:    8900,
			Close:  9100,
			Volume: 1000,
			Source: "yahoo",
		},
		{
			Ticker: "BBCA",
			Date:   time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			Open:   9100,
			High:   9300,
			Low:    9000,
			Close:  9250,
			Volume: 1500,
			Source: "yahoo",
		},
	}

	if err := repo.BulkUpsert(ctx, points); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].Close != 9100 {
		t.Errorf("got[0].Close = %v, want 9100", got[0].Close)
	}
	if got[1].Close != 9250 {
		t.Errorf("got[1].Close = %v, want 9250", got[1].Close)
	}
}

func TestPriceHistoryRepoBulkUpsertUpdatesExisting(t *testing.T) {
	repo, ctx := newPriceHistoryRepo(t)

	date := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	points := []stock.PricePoint{
		{Ticker: "BBCA", Date: date, Open: 9000, High: 9200, Low: 8900, Close: 9100, Volume: 1000, Source: "yahoo"},
	}
	if err := repo.BulkUpsert(ctx, points); err != nil {
		t.Fatalf("BulkUpsert() insert error = %v", err)
	}

	points[0].Close = 9500
	points[0].ID = "" // should get new ID but upsert on conflict
	if err := repo.BulkUpsert(ctx, points); err != nil {
		t.Fatalf("BulkUpsert() update error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].Close != 9500 {
		t.Errorf("Close = %v, want 9500", got[0].Close)
	}
}

func TestPriceHistoryRepoLatestDate(t *testing.T) {
	repo, ctx := newPriceHistoryRepo(t)

	// No data yet — should return zero time.
	latest, err := repo.LatestDate(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("LatestDate() error = %v", err)
	}
	if !latest.IsZero() {
		t.Errorf("LatestDate() = %v, want zero", latest)
	}

	points := []stock.PricePoint{
		{Ticker: "BBCA", Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Close: 9100, Source: "yahoo"},
		{Ticker: "BBCA", Date: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC), Close: 9500, Source: "yahoo"},
	}
	if err := repo.BulkUpsert(ctx, points); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	latest, err = repo.LatestDate(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("LatestDate() error = %v", err)
	}
	want := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	if !latest.Equal(want) {
		t.Errorf("LatestDate() = %v, want %v", latest, want)
	}
}

func TestPriceHistoryRepoDeleteByTicker(t *testing.T) {
	repo, ctx := newPriceHistoryRepo(t)

	points := []stock.PricePoint{
		{Ticker: "BBCA", Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Close: 9100, Source: "yahoo"},
		{Ticker: "TLKM", Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Close: 3200, Source: "yahoo"},
	}
	if err := repo.BulkUpsert(ctx, points); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	if err := repo.DeleteByTicker(ctx, "BBCA"); err != nil {
		t.Fatalf("DeleteByTicker() error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 after delete", len(got))
	}

	// TLKM should still exist.
	got, err = repo.GetByTicker(ctx, "TLKM", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker(TLKM) error = %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(got) = %d, want 1 (TLKM should remain)", len(got))
	}
}

func TestPriceHistoryRepoGetByTickerEmpty(t *testing.T) {
	repo, ctx := newPriceHistoryRepo(t)

	got, err := repo.GetByTicker(ctx, "NONEXIST", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestPriceHistoryRepoSourceIsolation(t *testing.T) {
	repo, ctx := newPriceHistoryRepo(t)

	date := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	points := []stock.PricePoint{
		{ID: shared.NewID(), Ticker: "BBCA", Date: date, Close: 9100, Source: "yahoo"},
		{ID: shared.NewID(), Ticker: "BBCA", Date: date, Close: 9200, Source: "other"},
	}
	if err := repo.BulkUpsert(ctx, points); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	yahoo, err := repo.GetByTicker(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("GetByTicker(yahoo) error = %v", err)
	}
	if len(yahoo) != 1 || yahoo[0].Close != 9100 {
		t.Errorf("yahoo data = %v, want 1 point with Close=9100", yahoo)
	}

	other, err := repo.GetByTicker(ctx, "BBCA", "other")
	if err != nil {
		t.Fatalf("GetByTicker(other) error = %v", err)
	}
	if len(other) != 1 || other[0].Close != 9200 {
		t.Errorf("other data = %v, want 1 point with Close=9200", other)
	}
}
