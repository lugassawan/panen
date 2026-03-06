package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/alert"
	"github.com/lugassawan/panen/backend/domain/shared"
)

func newTestAlert(ticker, metric string, sev alert.Severity) *alert.FundamentalAlert {
	return alert.NewFundamentalAlert(ticker, metric, sev, 20.0, 14.0, -30.0)
}

func TestAlertRepoCreateAndGetByTicker(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	a := newTestAlert("BBCA", "roe", alert.SeverityCritical)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	alerts, err := repo.GetByTicker(ctx, "BBCA")
	if err != nil {
		t.Fatalf("GetByTicker() error = %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("got %d alerts, want 1", len(alerts))
	}
	if alerts[0].Metric != "roe" {
		t.Errorf("metric = %q, want roe", alerts[0].Metric)
	}
	if alerts[0].Severity != alert.SeverityCritical {
		t.Errorf("severity = %q, want CRITICAL", alerts[0].Severity)
	}
}

func TestAlertRepoGetActive(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	a1 := newTestAlert("BBCA", "roe", alert.SeverityCritical)
	a2 := newTestAlert("BMRI", "der", alert.SeverityWarning)
	if err := repo.Create(ctx, a1); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, a2); err != nil {
		t.Fatal(err)
	}

	// Acknowledge one.
	if err := repo.Acknowledge(ctx, a1.ID); err != nil {
		t.Fatal(err)
	}

	active, err := repo.GetActive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 {
		t.Fatalf("got %d active alerts, want 1", len(active))
	}
	if active[0].ID != a2.ID {
		t.Errorf("expected a2 to be active, got %s", active[0].ID)
	}
}

func TestAlertRepoGetActiveByTicker(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	a1 := newTestAlert("BBCA", "roe", alert.SeverityCritical)
	a2 := newTestAlert("BMRI", "der", alert.SeverityWarning)
	if err := repo.Create(ctx, a1); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, a2); err != nil {
		t.Fatal(err)
	}

	active, err := repo.GetActiveByTicker(ctx, "BBCA")
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 {
		t.Fatalf("got %d, want 1", len(active))
	}
}

func TestAlertRepoAcknowledge(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	a := newTestAlert("BBCA", "roe", alert.SeverityCritical)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatal(err)
	}

	if err := repo.Acknowledge(ctx, a.ID); err != nil {
		t.Fatalf("Acknowledge() error = %v", err)
	}

	alerts, err := repo.GetByTicker(ctx, "BBCA")
	if err != nil {
		t.Fatal(err)
	}
	if alerts[0].Status != alert.AlertStatusAcknowledged {
		t.Errorf("status = %q, want ACKNOWLEDGED", alerts[0].Status)
	}
}

func TestAlertRepoAcknowledgeNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	err := repo.Acknowledge(ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAlertRepoResolve(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	a := newTestAlert("BBCA", "roe", alert.SeverityCritical)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatal(err)
	}

	if err := repo.Resolve(ctx, a.ID); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	alerts, err := repo.GetByTicker(ctx, "BBCA")
	if err != nil {
		t.Fatal(err)
	}
	if alerts[0].Status != alert.AlertStatusResolved {
		t.Errorf("status = %q, want RESOLVED", alerts[0].Status)
	}
	if alerts[0].ResolvedAt == nil {
		t.Error("expected ResolvedAt to be set")
	}
}

func TestAlertRepoCountActive(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	a1 := newTestAlert("BBCA", "roe", alert.SeverityCritical)
	a2 := newTestAlert("BMRI", "der", alert.SeverityWarning)
	if err := repo.Create(ctx, a1); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, a2); err != nil {
		t.Fatal(err)
	}

	count, err := repo.CountActive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	if err := repo.Resolve(ctx, a1.ID); err != nil {
		t.Fatal(err)
	}

	count, err = repo.CountActive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count after resolve = %d, want 1", count)
	}
}

func TestAlertRepoDeleteOlderThan(t *testing.T) {
	db := newTestDB(t)
	repo := NewAlertRepo(db)
	ctx := context.Background()

	a := newTestAlert("BBCA", "roe", alert.SeverityMinor)
	a.DetectedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatal(err)
	}

	deleted, err := repo.DeleteOlderThan(ctx, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}
}
