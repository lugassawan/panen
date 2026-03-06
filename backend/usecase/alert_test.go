package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/alert"
)

func TestAlertServiceHasCriticalAlert(t *testing.T) {
	repo := newMockAlertRepo()
	svc := NewAlertService(repo)
	ctx := context.Background()

	// No alerts → false.
	got, err := svc.HasCriticalAlert(ctx, "BBCA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false when no alerts exist")
	}

	// Add a WARNING alert → still false.
	warning := alert.NewFundamentalAlert("BBCA", "der", alert.SeverityWarning, 0.5, 1.1, 120.0)
	if err := repo.Create(ctx, warning); err != nil {
		t.Fatal(err)
	}
	got, err = svc.HasCriticalAlert(ctx, "BBCA")
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("expected false when only WARNING alert exists")
	}

	// Add a CRITICAL alert → true.
	critical := alert.NewFundamentalAlert("BBCA", "eps", alert.SeverityCritical, 500, -10, -102)
	if err := repo.Create(ctx, critical); err != nil {
		t.Fatal(err)
	}
	got, err = svc.HasCriticalAlert(ctx, "BBCA")
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected true when CRITICAL alert exists")
	}

	// Different ticker → false.
	got, err = svc.HasCriticalAlert(ctx, "BMRI")
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("expected false for different ticker")
	}
}

func TestAlertServiceGetActiveCount(t *testing.T) {
	repo := newMockAlertRepo()
	svc := NewAlertService(repo)
	ctx := context.Background()

	count, err := svc.GetActiveCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	a := alert.NewFundamentalAlert("BBCA", "pbv", alert.SeverityMinor, 3.0, 2.7, -10)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatal(err)
	}

	count, err = svc.GetActiveCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

// mockAlertRepo is a simple in-memory alert.Repository for usecase tests.
type mockAlertRepo struct {
	alerts []*alert.FundamentalAlert
}

func newMockAlertRepo() *mockAlertRepo {
	return &mockAlertRepo{}
}

func (m *mockAlertRepo) Create(_ context.Context, a *alert.FundamentalAlert) error {
	m.alerts = append(m.alerts, a)
	return nil
}

func (m *mockAlertRepo) GetByTicker(_ context.Context, ticker string) ([]*alert.FundamentalAlert, error) {
	var result []*alert.FundamentalAlert
	for _, a := range m.alerts {
		if a.Ticker == ticker {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAlertRepo) GetActive(_ context.Context) ([]*alert.FundamentalAlert, error) {
	var result []*alert.FundamentalAlert
	for _, a := range m.alerts {
		if a.Status == alert.AlertStatusActive {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAlertRepo) GetActiveByTicker(_ context.Context, ticker string) ([]*alert.FundamentalAlert, error) {
	var result []*alert.FundamentalAlert
	for _, a := range m.alerts {
		if a.Ticker == ticker && a.Status == alert.AlertStatusActive {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAlertRepo) Acknowledge(_ context.Context, id string) error {
	for _, a := range m.alerts {
		if a.ID == id {
			a.Status = alert.AlertStatusAcknowledged
			return nil
		}
	}
	return nil
}

func (m *mockAlertRepo) Resolve(_ context.Context, id string) error {
	for _, a := range m.alerts {
		if a.ID == id {
			a.Status = alert.AlertStatusResolved
			return nil
		}
	}
	return nil
}

func (m *mockAlertRepo) CountActive(_ context.Context) (int, error) {
	count := 0
	for _, a := range m.alerts {
		if a.Status == alert.AlertStatusActive {
			count++
		}
	}
	return count, nil
}

func (m *mockAlertRepo) DeleteOlderThan(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}
