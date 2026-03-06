package dividend

import (
	"testing"
	"time"
)

func TestProjectUpcoming(t *testing.T) {
	// Historical events: paid in March and September for 2022-2024
	events := []DividendEvent{
		{Ticker: "BBCA", ExDate: time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50},
		{Ticker: "BBCA", ExDate: time.Date(2022, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 40},
		{Ticker: "BBCA", ExDate: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 55},
		{Ticker: "BBCA", ExDate: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 45},
		{Ticker: "BBCA", ExDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 60},
		{Ticker: "BBCA", ExDate: time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 50},
	}

	projections := ProjectUpcoming(events, "BBCA")

	// Should project 1-2 upcoming months in the next 12 months
	if len(projections) == 0 {
		t.Fatal("expected projections, got 0")
	}

	now := time.Now().UTC()
	horizon := now.AddDate(1, 0, 0)
	for _, p := range projections {
		if p.Ticker != "BBCA" {
			t.Errorf("ticker = %s, want BBCA", p.Ticker)
		}
		if !p.IsProjection {
			t.Error("expected IsProjection = true")
		}
		if p.ExpectedExDate.Before(now) {
			t.Errorf("projected date %v is in the past", p.ExpectedExDate)
		}
		if p.ExpectedExDate.After(horizon) {
			t.Errorf("projected date %v is beyond 12-month horizon", p.ExpectedExDate)
		}
		if p.ExpectedAmount <= 0 {
			t.Errorf("expected positive amount, got %v", p.ExpectedAmount)
		}
	}
}

func TestProjectUpcomingEmpty(t *testing.T) {
	result := ProjectUpcoming(nil, "BBCA")
	if result != nil {
		t.Errorf("got %v, want nil", result)
	}
}
