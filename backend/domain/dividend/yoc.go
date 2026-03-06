package dividend

import "time"

// DeriveAnnualDPS estimates the annual dividend per share from price and
// dividend yield percentage. Returns 0 if inputs are non-positive.
func DeriveAnnualDPS(price, dividendYieldPct float64) float64 {
	if price <= 0 || dividendYieldPct <= 0 {
		return 0
	}
	return price * dividendYieldPct / 100
}

// YieldOnCost computes the yield on cost as a percentage.
// Returns 0 if avgBuyPrice is non-positive.
func YieldOnCost(annualDPS, avgBuyPrice float64) float64 {
	if avgBuyPrice <= 0 || annualDPS <= 0 {
		return 0
	}
	return annualDPS / avgBuyPrice * 100
}

// ProjectedYoC computes the expected yield on cost after averaging up
// by purchasing newLots at newPrice, given the current position.
// Returns 0 if the resulting average price would be non-positive.
func ProjectedYoC(annualDPS, avgBuyPrice float64, currentLots int, newPrice float64, newLots int) float64 {
	if currentLots < 0 || newLots <= 0 || avgBuyPrice <= 0 || newPrice <= 0 {
		return 0
	}
	totalShares := float64(currentLots+newLots) * 100 // IDX: 1 lot = 100 shares
	currentShares := float64(currentLots) * 100
	newShares := float64(newLots) * 100
	newAvg := (currentShares*avgBuyPrice + newShares*newPrice) / totalShares
	if newAvg <= 0 {
		return 0
	}
	return annualDPS / newAvg * 100
}

// YoCPoint represents yield on cost at a specific point in time.
type YoCPoint struct {
	Date time.Time
	YoC  float64
}

// YoCProgression computes historical yield on cost at each dividend payment date.
// It sums trailing-12-month DPS at each event and divides by avgBuyPrice.
func YoCProgression(events []DividendEvent, avgBuyPrice float64) []YoCPoint {
	if avgBuyPrice <= 0 || len(events) == 0 {
		return nil
	}

	var points []YoCPoint
	for i, e := range events {
		if e.Amount <= 0 {
			continue
		}
		// Sum trailing 12-month DPS up to and including this event
		cutoff := e.ExDate.AddDate(-1, 0, 0)
		var trailing float64
		for j := 0; j <= i; j++ {
			if events[j].ExDate.After(cutoff) && events[j].Amount > 0 {
				trailing += events[j].Amount
			}
		}
		yoc := trailing / avgBuyPrice * 100
		points = append(points, YoCPoint{Date: e.ExDate, YoC: yoc})
	}
	return points
}

// PortfolioYieldItem is used to compute portfolio-level weighted yield.
type PortfolioYieldItem struct {
	PositionValue float64 // currentPrice * lots * 100
	AnnualDPS     float64
	Lots          int
}

// PortfolioYield computes the weighted portfolio-level dividend yield.
// The weight of each holding is its position value relative to the total.
func PortfolioYield(items []PortfolioYieldItem) float64 {
	var totalValue float64
	var totalDividendIncome float64
	for _, item := range items {
		if item.PositionValue <= 0 {
			continue
		}
		totalValue += item.PositionValue
		totalDividendIncome += item.AnnualDPS * float64(item.Lots) * 100
	}
	if totalValue <= 0 {
		return 0
	}
	return totalDividendIncome / totalValue * 100
}
