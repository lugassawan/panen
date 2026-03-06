package dividend

import (
	"testing"
	"time"
)

func TestAnnualDividendIncome(t *testing.T) {
	now := time.Now().UTC()
	events := []DividendEvent{
		{ExDate: now.AddDate(0, -6, 0), Amount: 50},
		{ExDate: now.AddDate(0, -3, 0), Amount: 60},
		{ExDate: now.AddDate(-2, 0, 0), Amount: 100}, // outside 12-month window
	}

	// 10 lots * 100 shares = 1000 shares
	// income = (50 + 60) * 1000 = 110000
	got := AnnualDividendIncome(events, 10)
	if !almostEqual(got, 110000) {
		t.Errorf("AnnualDividendIncome() = %v, want 110000", got)
	}
}

func TestAnnualDividendIncomeZeroLots(t *testing.T) {
	events := []DividendEvent{
		{ExDate: time.Now().UTC().AddDate(0, -1, 0), Amount: 50},
	}
	got := AnnualDividendIncome(events, 0)
	if got != 0 {
		t.Errorf("AnnualDividendIncome(lots=0) = %v, want 0", got)
	}
}

func TestMonthlyDividendIncome(t *testing.T) {
	now := time.Now().UTC()
	events := []DividendEvent{
		{ExDate: now.AddDate(0, -6, 0), Amount: 50},
		{ExDate: now.AddDate(0, -3, 0), Amount: 60},
	}

	months := MonthlyDividendIncome(events, 5) // 5 lots = 500 shares
	if len(months) == 0 {
		t.Fatal("got 0 months, want > 0")
	}

	var total float64
	for _, m := range months {
		total += m.Amount
	}
	want := (50 + 60) * 500.0
	if !almostEqual(total, want) {
		t.Errorf("total monthly income = %v, want %v", total, want)
	}
}

func TestMonthlyDividendIncomeNegativeLots(t *testing.T) {
	events := []DividendEvent{
		{ExDate: time.Now().UTC().AddDate(0, -1, 0), Amount: 50},
	}
	result := MonthlyDividendIncome(events, -1)
	if result != nil {
		t.Errorf("got %v, want nil", result)
	}
}
