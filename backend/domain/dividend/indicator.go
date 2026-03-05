package dividend

// IndicatorInput holds all data needed to determine a dividend indicator.
type IndicatorInput struct {
	HasHolding     bool
	Price          float64
	EntryPrice     float64
	ExitTarget     float64
	DividendYield  float64
	PayoutRatio    float64
	PositionPct    float64
	MinDY          float64
	MaxPayoutRatio float64
	MaxPositionPct float64
}

// DetermineIndicator classifies a stock into one of the four dividend indicators.
//
// Priority order:
//  1. OVERVALUED if price >= exit target OR payout ratio exceeds max threshold
//  2. BUY_ZONE if no holding AND price <= entry AND DY >= min
//  3. AVERAGE_UP if has holding AND DY >= min AND payout sustainable AND position weight OK
//  4. HOLD (default for existing holdings)
func DetermineIndicator(input IndicatorInput) Indicator {
	if input.ExitTarget > 0 && input.Price >= input.ExitTarget {
		return IndicatorOvervalued
	}
	if input.MaxPayoutRatio > 0 && input.PayoutRatio > input.MaxPayoutRatio {
		return IndicatorOvervalued
	}

	if !input.HasHolding {
		if input.EntryPrice > 0 && input.Price <= input.EntryPrice && input.DividendYield >= input.MinDY {
			return IndicatorBuyZone
		}
		return IndicatorHold
	}

	if input.DividendYield >= input.MinDY &&
		input.PayoutRatio <= input.MaxPayoutRatio &&
		input.PositionPct < input.MaxPositionPct {
		return IndicatorAverageUp
	}

	return IndicatorHold
}
