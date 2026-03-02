package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/shared"
	"github.com/lugassawan/panen/backend/internal/domain/stock"
)

func newTestStockData(now time.Time) *stock.Data {
	return &stock.Data{
		ID:            shared.NewID(),
		Ticker:        "BBCA",
		Price:         8500,
		High52Week:    9200,
		Low52Week:     7800,
		EPS:           500,
		BVPS:          3200,
		ROE:           18.5,
		DER:           5.2,
		PBV:           2.65,
		PER:           17,
		DividendYield: 2.8,
		PayoutRatio:   50,
		FetchedAt:     now,
		Source:        "idx",
	}
}

func TestStockDataRepoUpsertAndGetByTicker(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	repo := NewStockDataRepo(conn)

	d := newTestStockData(now)
	if err := repo.Upsert(ctx, d); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBCA")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if got.Price != 8500 {
		t.Errorf("Price = %f, want 8500", got.Price)
	}
	if got.Source != "idx" {
		t.Errorf("Source = %q, want %q", got.Source, "idx")
	}
}

func TestStockDataRepoUpsertOverwrites(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	repo := NewStockDataRepo(conn)

	d1 := &stock.Data{
		ID: shared.NewID(), Ticker: "BBRI", Price: 4500,
		FetchedAt: now, Source: "idx",
	}
	if err := repo.Upsert(ctx, d1); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// Upsert with a NEW id but same (ticker, source) should update.
	d2 := &stock.Data{
		ID: shared.NewID(), Ticker: "BBRI", Price: 4600,
		FetchedAt: now.Add(time.Hour), Source: "idx",
	}
	if err := repo.Upsert(ctx, d2); err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}

	got, err := repo.GetByTicker(ctx, "BBRI")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if got.Price != 4600 {
		t.Errorf("Price = %f, want 4600 after upsert", got.Price)
	}
}

func TestStockDataRepoGetByTickerNotFound(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	repo := NewStockDataRepo(conn)

	_, err := repo.GetByTicker(ctx, "NONEXISTENT")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByTicker() error = %v, want ErrNotFound", err)
	}
}

func TestStockDataRepoGetByTickerAndSource(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	repo := NewStockDataRepo(conn)

	d := &stock.Data{
		ID: shared.NewID(), Ticker: "TLKM", Price: 3500,
		FetchedAt: now, Source: "stockbit",
	}
	if err := repo.Upsert(ctx, d); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := repo.GetByTickerAndSource(ctx, "TLKM", "stockbit")
	if err != nil {
		t.Fatalf("GetByTickerAndSource() error = %v", err)
	}
	if got.Source != "stockbit" {
		t.Errorf("Source = %q, want %q", got.Source, "stockbit")
	}

	_, err = repo.GetByTickerAndSource(ctx, "TLKM", "other")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("wrong source error = %v, want ErrNotFound", err)
	}
}

func TestStockDataRepoDeleteOlderThan(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	repo := NewStockDataRepo(conn)

	old := &stock.Data{
		ID: shared.NewID(), Ticker: "OLD", Price: 1000,
		FetchedAt: now.Add(-48 * time.Hour), Source: "test",
	}
	if err := repo.Upsert(ctx, old); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	cutoff := now.Add(-24 * time.Hour)
	n, err := repo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}
	if n < 1 {
		t.Errorf("DeleteOlderThan() deleted %d rows, want >= 1", n)
	}

	_, err = repo.GetByTicker(ctx, "OLD")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByTicker() after delete error = %v, want ErrNotFound", err)
	}
}
