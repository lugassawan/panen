package dividend

import (
	"sort"
	"time"
)

// MonthlyIncome holds dividend income for a single month.
type MonthlyIncome struct {
	Month  int
	Amount float64
}

// AnnualDividendIncome computes total dividend income for the last 12 months
// given the number of lots held (1 lot = 100 shares in IDX).
func AnnualDividendIncome(events []DividendEvent, lots int) float64 {
	if lots <= 0 {
		return 0
	}

	cutoff := time.Now().UTC().AddDate(-1, 0, 0)
	shares := float64(lots) * 100

	var total float64
	for _, e := range events {
		if e.ExDate.After(cutoff) && e.Amount > 0 {
			total += e.Amount * shares
		}
	}
	return total
}

// MonthlyDividendIncome buckets dividend events from the last 12 months
// into monthly income amounts (DPS * lots * 100).
func MonthlyDividendIncome(events []DividendEvent, lots int) []MonthlyIncome {
	if lots <= 0 {
		return nil
	}

	cutoff := time.Now().UTC().AddDate(-1, 0, 0)
	shares := float64(lots) * 100
	monthMap := make(map[int]float64)

	for _, e := range events {
		if e.ExDate.After(cutoff) && e.Amount > 0 {
			monthMap[int(e.ExDate.Month())] += e.Amount * shares
		}
	}

	result := make([]MonthlyIncome, 0, len(monthMap))
	for m, amt := range monthMap {
		result = append(result, MonthlyIncome{Month: m, Amount: amt})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Month < result[j].Month })
	return result
}
