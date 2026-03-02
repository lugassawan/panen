package valuation

// RiskProfile defines the risk tolerance level for valuation calculations.
type RiskProfile string

const (
	RiskConservative RiskProfile = "CONSERVATIVE"
	RiskModerate     RiskProfile = "MODERATE"
	RiskAggressive   RiskProfile = "AGGRESSIVE"
)

// Verdict indicates whether a stock is undervalued, fair, or overvalued.
type Verdict string

const (
	VerdictUndervalued Verdict = "UNDERVALUED"
	VerdictFair        Verdict = "FAIR"
	VerdictOvervalued  Verdict = "OVERVALUED"
)
