package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
)

func TestSnapshotRepoInsertAndGetLatest(t *testing.T) {
	db := newTestDB(t)
	repo := NewSnapshotRepo(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	data := &stock.Data{
		Ticker:        "BBCA",
		Price:         8000,
		EPS:           500,
		BVPS:          2500,
		ROE:           20,
		DER:           0.5,
		PBV:           3.2,
		PER:           16,
		DividendYield: 3.0,
		PayoutRatio:   40,
		Source:        "yahoo",
		FetchedAt:     now,
	}

	if err := repo.Insert(ctx, data); err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	got, err := repo.GetLatest(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatalf("GetLatest() error = %v", err)
	}
	if got.Ticker != "BBCA" {
		t.Errorf("ticker = %q, want BBCA", got.Ticker)
	}
	if got.ROE != 20 {
		t.Errorf("ROE = %v, want 20", got.ROE)
	}
}

func TestSnapshotRepoGetLatestNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewSnapshotRepo(db)
	ctx := context.Background()

	_, err := repo.GetLatest(ctx, "NONEXIST", "yahoo")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSnapshotRepoGetLatestReturnsNewest(t *testing.T) {
	db := newTestDB(t)
	repo := NewSnapshotRepo(db)
	ctx := context.Background()

	old := &stock.Data{
		Ticker: "BBCA", Price: 7000, ROE: 18, Source: "yahoo",
		FetchedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	newer := &stock.Data{
		Ticker: "BBCA", Price: 8000, ROE: 20, Source: "yahoo",
		FetchedAt: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := repo.Insert(ctx, old); err != nil {
		t.Fatal(err)
	}
	if err := repo.Insert(ctx, newer); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetLatest(ctx, "BBCA", "yahoo")
	if err != nil {
		t.Fatal(err)
	}
	if got.Price != 8000 {
		t.Errorf("Price = %v, want 8000 (latest)", got.Price)
	}
}

func TestSnapshotRepoCleanup(t *testing.T) {
	db := newTestDB(t)
	repo := NewSnapshotRepo(db)
	ctx := context.Background()

	// Insert 5 snapshots.
	for i := range 5 {
		d := &stock.Data{
			ID:        shared.NewID(),
			Ticker:    "BBCA",
			Price:     float64(7000 + i*100),
			Source:    "yahoo",
			FetchedAt: time.Date(2025, 1, 1+i, 0, 0, 0, 0, time.UTC),
		}
		if err := repo.Insert(ctx, d); err != nil {
			t.Fatal(err)
		}
	}

	// Keep only 2.
	if err := repo.Cleanup(ctx, "BBCA", 2); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	// Verify count by inserting a dummy query.
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM financial_snapshots WHERE ticker = ?", "BBCA").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("snapshot count = %d, want 2", count)
	}
}
