package portfolio

// Mode defines the investment strategy of a portfolio.
type Mode string

const (
	ModeValue    Mode = "VALUE"
	ModeDividend Mode = "DIVIDEND"
)

// RiskProfile defines the risk tolerance level of a portfolio.
type RiskProfile string

const (
	RiskProfileConservative RiskProfile = "CONSERVATIVE"
	RiskProfileModerate     RiskProfile = "MODERATE"
	RiskProfileAggressive   RiskProfile = "AGGRESSIVE"
)
