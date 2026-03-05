package trailingstop

import (
	"fmt"

	"github.com/lugassawan/panen/backend/domain/portfolio"
)

// StopPercentage returns the trailing stop percentage for the given risk profile.
// Conservative: 9%, Moderate: 13.5%, Aggressive: 21.5%.
func StopPercentage(riskProfile portfolio.RiskProfile) (float64, error) {
	switch riskProfile {
	case portfolio.RiskProfileConservative:
		return 9.0, nil
	case portfolio.RiskProfileModerate:
		return 13.5, nil
	case portfolio.RiskProfileAggressive:
		return 21.5, nil
	default:
		return 0, fmt.Errorf("invalid risk profile: %s", riskProfile)
	}
}

// StopPrice computes the absolute stop price given a peak and stop percentage.
func StopPrice(peakPrice, stopPct float64) float64 {
	return peakPrice * (1 - stopPct/100)
}

// IsTriggered returns true when the current price is at or below the stop price.
func IsTriggered(currentPrice, stopPrice float64) bool {
	return currentPrice <= stopPrice
}

// UpdatePeak returns the higher of the current peak and the current price.
func UpdatePeak(currentPeak, currentPrice float64) float64 {
	if currentPrice > currentPeak {
		return currentPrice
	}
	return currentPeak
}

// EvaluateFundamentals checks for fundamental deterioration signals.
// Returns a list of exit criteria with their triggered state.
func EvaluateFundamentals(roe, der, eps float64) []FundamentalExit {
	return []FundamentalExit{
		{
			Key:       "roe_low",
			Label:     "ROE below 10%",
			Detail:    fmt.Sprintf("ROE is %.1f%%", roe),
			Triggered: roe < 10,
		},
		{
			Key:       "der_high",
			Label:     "DER above 1.5x",
			Detail:    fmt.Sprintf("DER is %.2fx", der),
			Triggered: der > 1.5,
		},
		{
			Key:       "eps_negative",
			Label:     "EPS zero or negative",
			Detail:    fmt.Sprintf("EPS is %.2f", eps),
			Triggered: eps <= 0,
		},
	}
}
