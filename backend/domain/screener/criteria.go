package screener

import "github.com/lugassawan/panen/backend/domain/valuation"

// Criteria holds the fundamental quality thresholds for screening stocks.
type Criteria struct {
	MinROE float64
	MaxDER float64
}

// CriteriaForRisk returns screening criteria based on the investor's risk profile.
func CriteriaForRisk(rp valuation.RiskProfile) Criteria {
	switch rp {
	case valuation.RiskConservative:
		return Criteria{MinROE: 15, MaxDER: 0.8}
	case valuation.RiskModerate:
		return Criteria{MinROE: 12, MaxDER: 1.0}
	case valuation.RiskAggressive:
		return Criteria{MinROE: 8, MaxDER: 1.5}
	default:
		return Criteria{MinROE: 12, MaxDER: 1.0}
	}
}
