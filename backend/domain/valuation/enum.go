package valuation

import "fmt"

// RiskProfile defines the risk tolerance level for valuation calculations.
type RiskProfile string

const (
	RiskConservative RiskProfile = "CONSERVATIVE"
	RiskModerate     RiskProfile = "MODERATE"
	RiskAggressive   RiskProfile = "AGGRESSIVE"
)

// ParseRiskProfile converts a string to a RiskProfile enum value.
func ParseRiskProfile(s string) (RiskProfile, error) {
	switch s {
	case "CONSERVATIVE":
		return RiskConservative, nil
	case "MODERATE":
		return RiskModerate, nil
	case "AGGRESSIVE":
		return RiskAggressive, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidRisk, s)
	}
}

// Verdict indicates whether a stock is undervalued, fair, or overvalued.
type Verdict string

const (
	VerdictUndervalued Verdict = "UNDERVALUED"
	VerdictFair        Verdict = "FAIR"
	VerdictOvervalued  Verdict = "OVERVALUED"
)
