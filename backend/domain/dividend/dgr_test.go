package dividend

import (
	"testing"
	"time"
)

func TestAggregateAnnualDPS(t *testing.T) {
	events := []DividendEvent{
		{ExDate: time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50},
		{ExDate: time.Date(2022, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 50},
		{ExDate: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 60},
		{ExDate: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 60},
		{ExDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 75},
	}

	annuals := AggregateAnnualDPS(events)
	if len(annuals) != 3 {
		t.Fatalf("got %d annuals, want 3", len(annuals))
	}

	want := []AnnualDividend{
		{Year: 2022, TotalDPS: 100},
		{Year: 2023, TotalDPS: 120},
		{Year: 2024, TotalDPS: 75},
	}
	for i, a := range annuals {
		if a.Year != want[i].Year || !almostEqual(a.TotalDPS, want[i].TotalDPS) {
			t.Errorf("annuals[%d] = %+v, want %+v", i, a, want[i])
		}
	}
}

func TestAggregateAnnualDPSEmpty(t *testing.T) {
	result := AggregateAnnualDPS(nil)
	if len(result) != 0 {
		t.Errorf("got %d annuals, want 0", len(result))
	}
}

func TestAggregateAnnualDPSSkipsNegative(t *testing.T) {
	events := []DividendEvent{
		{ExDate: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC), Amount: -10},
		{ExDate: time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC), Amount: 50},
	}
	annuals := AggregateAnnualDPS(events)
	if len(annuals) != 1 || !almostEqual(annuals[0].TotalDPS, 50) {
		t.Errorf("got %+v, want [{2023 50}]", annuals)
	}
}

func TestCalculateDGR(t *testing.T) {
	annuals := []AnnualDividend{
		{Year: 2021, TotalDPS: 100},
		{Year: 2022, TotalDPS: 120},
		{Year: 2023, TotalDPS: 108},
	}

	results := CalculateDGR(annuals)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}

	if results[0].GrowthPct != 0 {
		t.Errorf("first year growth = %v, want 0", results[0].GrowthPct)
	}
	if !almostEqual(results[1].GrowthPct, 20) {
		t.Errorf("2022 growth = %v, want 20", results[1].GrowthPct)
	}
	if !almostEqual(results[2].GrowthPct, -10) {
		t.Errorf("2023 growth = %v, want -10", results[2].GrowthPct)
	}
}

func TestCalculateDGREmpty(t *testing.T) {
	result := CalculateDGR(nil)
	if result != nil {
		t.Errorf("got %v, want nil", result)
	}
}
