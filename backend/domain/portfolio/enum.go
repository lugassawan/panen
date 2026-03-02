package portfolio

import (
	"errors"
	"fmt"
)

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

var (
	ErrInvalidMode = errors.New("invalid portfolio mode")
	ErrInvalidRisk = errors.New("invalid risk profile")
)

// ParseMode converts a string to a Mode enum value.
func ParseMode(s string) (Mode, error) {
	switch s {
	case "VALUE":
		return ModeValue, nil
	case "DIVIDEND":
		return ModeDividend, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidMode, s)
	}
}

// ParseRiskProfile converts a string to a RiskProfile enum value.
func ParseRiskProfile(s string) (RiskProfile, error) {
	switch s {
	case "CONSERVATIVE":
		return RiskProfileConservative, nil
	case "MODERATE":
		return RiskProfileModerate, nil
	case "AGGRESSIVE":
		return RiskProfileAggressive, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidRisk, s)
	}
}
