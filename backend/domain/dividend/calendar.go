package dividend

import (
	"sort"
	"time"
)

// ProjectedDividend represents an upcoming dividend event (actual or projected).
type ProjectedDividend struct {
	Ticker         string
	ExpectedExDate time.Time
	ExpectedAmount float64
	IsProjection   bool
}

type monthInfo struct {
	latestYear   int
	latestAmount float64
	typicalDay   int
}

// ProjectUpcoming uses historical dividend patterns to project upcoming dividends
// for the next 12 months. It identifies which months the stock typically pays
// dividends and projects future payments using the most recent amount for each month.
func ProjectUpcoming(events []DividendEvent, ticker string) []ProjectedDividend {
	if len(events) == 0 {
		return nil
	}

	now := time.Now().UTC()
	horizon := now.AddDate(1, 0, 0)

	monthData := make(map[time.Month]*monthInfo)

	for _, e := range events {
		if e.Amount <= 0 {
			continue
		}
		m := e.ExDate.Month()
		info, ok := monthData[m]
		if !ok {
			monthData[m] = &monthInfo{
				latestYear:   e.ExDate.Year(),
				latestAmount: e.Amount,
				typicalDay:   e.ExDate.Day(),
			}
			continue
		}
		if e.ExDate.Year() > info.latestYear {
			info.latestYear = e.ExDate.Year()
			info.latestAmount = e.Amount
			info.typicalDay = e.ExDate.Day()
		}
	}

	var projections []ProjectedDividend
	for m := time.Month(1); m <= 12; m++ {
		info, ok := monthData[m]
		if !ok {
			continue
		}

		// Project for the next occurrence of this month
		year := now.Year()
		projected := time.Date(year, m, info.typicalDay, 0, 0, 0, 0, time.UTC)
		if projected.Before(now) {
			projected = time.Date(year+1, m, info.typicalDay, 0, 0, 0, 0, time.UTC)
		}
		if projected.After(horizon) {
			continue
		}

		projections = append(projections, ProjectedDividend{
			Ticker:         ticker,
			ExpectedExDate: projected,
			ExpectedAmount: info.latestAmount,
			IsProjection:   true,
		})
	}

	sort.Slice(projections, func(i, j int) bool {
		return projections[i].ExpectedExDate.Before(projections[j].ExpectedExDate)
	})
	return projections
}
