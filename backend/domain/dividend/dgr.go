package dividend

import "sort"

// AnnualDividend holds the total DPS paid in a calendar year.
type AnnualDividend struct {
	Year     int
	TotalDPS float64
}

// DGRResult holds dividend growth rate data for a single year.
type DGRResult struct {
	Year      int
	DPS       float64
	GrowthPct float64
}

// AggregateAnnualDPS groups dividend events by year and sums the DPS per year.
// Returns results sorted by year ascending.
func AggregateAnnualDPS(events []DividendEvent) []AnnualDividend {
	yearMap := make(map[int]float64)
	for _, e := range events {
		if e.Amount <= 0 {
			continue
		}
		yearMap[e.ExDate.Year()] += e.Amount
	}

	result := make([]AnnualDividend, 0, len(yearMap))
	for year, total := range yearMap {
		result = append(result, AnnualDividend{Year: year, TotalDPS: total})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Year < result[j].Year })
	return result
}

// CalculateDGR computes year-over-year dividend growth rates.
// The first year has GrowthPct = 0 (no prior year to compare).
func CalculateDGR(annuals []AnnualDividend) []DGRResult {
	if len(annuals) == 0 {
		return nil
	}

	results := make([]DGRResult, len(annuals))
	results[0] = DGRResult{Year: annuals[0].Year, DPS: annuals[0].TotalDPS}
	for i := 1; i < len(annuals); i++ {
		cur := annuals[i]
		prev := annuals[i-1] //nolint:gosec // i starts at 1, so i-1 is always valid
		results[i] = DGRResult{Year: cur.Year, DPS: cur.TotalDPS}
		if prev.TotalDPS > 0 {
			results[i].GrowthPct = (cur.TotalDPS - prev.TotalDPS) / prev.TotalDPS * 100
		}
	}
	return results
}
