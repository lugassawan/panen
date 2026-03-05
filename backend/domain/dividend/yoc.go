package dividend

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
