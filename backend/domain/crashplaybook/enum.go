package crashplaybook

import "fmt"

// MarketCondition represents the current state of the broad market.
type MarketCondition string

const (
	MarketNormal     MarketCondition = "NORMAL"
	MarketElevated   MarketCondition = "ELEVATED"
	MarketCorrection MarketCondition = "CORRECTION"
	MarketCrash      MarketCondition = "CRASH"
	MarketRecovery   MarketCondition = "RECOVERY"
)

// CrashLevel represents a pre-calculated crash response tier.
type CrashLevel string

const (
	LevelNormalDip CrashLevel = "NORMAL_DIP"
	LevelCrash     CrashLevel = "CRASH"
	LevelExtreme   CrashLevel = "EXTREME"
)

// DiagnosticSignal is the result of a falling knife diagnostic evaluation.
type DiagnosticSignal string

const (
	SignalOpportunity  DiagnosticSignal = "OPPORTUNITY"
	SignalFallingKnife DiagnosticSignal = "FALLING_KNIFE"
	SignalInconclusive DiagnosticSignal = "INCONCLUSIVE"
)

// ParseMarketCondition converts a string to a MarketCondition enum value.
func ParseMarketCondition(s string) (MarketCondition, error) {
	switch s {
	case "NORMAL":
		return MarketNormal, nil
	case "ELEVATED":
		return MarketElevated, nil
	case "CORRECTION":
		return MarketCorrection, nil
	case "CRASH":
		return MarketCrash, nil
	case "RECOVERY":
		return MarketRecovery, nil
	default:
		return "", fmt.Errorf("unknown market condition: %s", s)
	}
}

// ParseCrashLevel converts a string to a CrashLevel enum value.
func ParseCrashLevel(s string) (CrashLevel, error) {
	switch s {
	case "NORMAL_DIP":
		return LevelNormalDip, nil
	case "CRASH":
		return LevelCrash, nil
	case "EXTREME":
		return LevelExtreme, nil
	default:
		return "", fmt.Errorf("unknown crash level: %s", s)
	}
}
