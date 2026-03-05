package crashplaybook

// Drawdown thresholds for market condition detection.
const (
	thresholdElevated   = -5.0
	thresholdCorrection = -10.0
	thresholdCrash      = -20.0
	thresholdRecovery   = -10.0
)

// DetectMarketCondition determines the market condition from IHSG price and peak.
// previousCondition is used to detect recovery (must have been in Crash or Correction).
func DetectMarketCondition(price, peak float64, previousCondition MarketCondition) MarketCondition {
	if peak <= 0 {
		return MarketNormal
	}

	drawdown := ((price - peak) / peak) * 100

	if isRecovering(drawdown, previousCondition) {
		return MarketRecovery
	}

	if drawdown <= thresholdCrash {
		return MarketCrash
	}
	if drawdown <= thresholdCorrection {
		return MarketCorrection
	}
	if drawdown <= thresholdElevated {
		return MarketElevated
	}
	return MarketNormal
}

// DrawdownPct calculates the percentage drawdown from peak.
func DrawdownPct(price, peak float64) float64 {
	if peak <= 0 {
		return 0
	}
	return ((price - peak) / peak) * 100
}

func isRecovering(drawdown float64, prev MarketCondition) bool {
	wasCrashOrCorrection := prev == MarketCrash || prev == MarketCorrection || prev == MarketRecovery
	return wasCrashOrCorrection && drawdown > thresholdRecovery && drawdown <= thresholdElevated
}
