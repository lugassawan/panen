package checklist

import "github.com/lugassawan/panen/backend/domain/portfolio"

// CheckDefinition describes a single check item in a checklist.
type CheckDefinition struct {
	Key   string
	Label string
	Type  CheckType
}

// Thresholds holds risk-profile-dependent threshold values for auto-checks.
type Thresholds struct {
	MinROE         float64
	MaxDER         float64
	MaxPositionPct float64
	MinDY          float64
	MaxPayoutRatio float64
}

var thresholdsByRisk = map[portfolio.RiskProfile]Thresholds{
	portfolio.RiskProfileConservative: {
		MinROE:         15,
		MaxDER:         0.8,
		MaxPositionPct: 10,
		MinDY:          5,
		MaxPayoutRatio: 60,
	},
	portfolio.RiskProfileModerate: {
		MinROE:         12,
		MaxDER:         1.0,
		MaxPositionPct: 20,
		MinDY:          3,
		MaxPayoutRatio: 75,
	},
	portfolio.RiskProfileAggressive: {
		MinROE:         8,
		MaxDER:         1.5,
		MaxPositionPct: 35,
		MinDY:          2,
		MaxPayoutRatio: 90,
	},
}

var autoCheckDefs = map[ActionType][]CheckDefinition{
	ActionBuy: {
		{Key: "roe_above_min", Label: "ROE above minimum threshold", Type: CheckTypeAuto},
		{Key: "der_below_max", Label: "DER below maximum threshold", Type: CheckTypeAuto},
		{Key: "price_below_entry", Label: "Price below entry target", Type: CheckTypeAuto},
		{Key: "position_weight", Label: "Position weight within limit", Type: CheckTypeAuto},
	},
	ActionAverageDown: {
		{Key: "roe_above_min", Label: "ROE above minimum threshold", Type: CheckTypeAuto},
		{Key: "der_below_max", Label: "DER below maximum threshold", Type: CheckTypeAuto},
		{Key: "price_below_entry", Label: "Price below entry target", Type: CheckTypeAuto},
		{Key: "current_loss", Label: "Current position at a loss", Type: CheckTypeAuto},
		{Key: "new_avg_price", Label: "New average price is lower", Type: CheckTypeAuto},
	},
	ActionAverageUp: {
		{Key: "dy_above_min", Label: "Dividend yield above minimum", Type: CheckTypeAuto},
		{Key: "payout_sustainable", Label: "Payout ratio is sustainable", Type: CheckTypeAuto},
		{Key: "price_below_entry", Label: "Price below entry target", Type: CheckTypeAuto},
		{Key: "position_weight", Label: "Position weight within limit", Type: CheckTypeAuto},
	},
	ActionSellExit: {
		{Key: "price_above_exit", Label: "Price above exit target", Type: CheckTypeAuto},
		{Key: "capital_gain", Label: "Capital gain is positive", Type: CheckTypeAuto},
	},
	ActionSellStop: {
		{Key: "fundamentals_stable", Label: "Fundamentals remain stable", Type: CheckTypeAuto},
		{Key: "capital_gain", Label: "Capital gain is positive", Type: CheckTypeAuto},
	},
	ActionHold: {
		{Key: "fundamentals_stable", Label: "Fundamentals remain stable", Type: CheckTypeAuto},
		{Key: "dividend_maintained", Label: "Dividend is maintained", Type: CheckTypeAuto},
	},
}

var manualCheckDefs = map[ActionType][]CheckDefinition{
	ActionBuy: {
		{Key: "no_negative_news", Label: "No negative news or red flags", Type: CheckTypeManual},
		{Key: "thesis_still_valid", Label: "Investment thesis still valid", Type: CheckTypeManual},
		{Key: "reviewed_financials", Label: "Reviewed latest financials", Type: CheckTypeManual},
	},
	ActionAverageDown: {
		{Key: "confirmed_not_value_trap", Label: "Confirmed not a value trap", Type: CheckTypeManual},
		{Key: "reviewed_downside_catalyst", Label: "Reviewed downside catalyst", Type: CheckTypeManual},
	},
	ActionAverageUp: {
		{Key: "dividend_track_record", Label: "Dividend track record is solid", Type: CheckTypeManual},
		{Key: "no_payout_cut_risk", Label: "No payout cut risk", Type: CheckTypeManual},
		{Key: "dividend_growth_positive", Label: "Dividend has been growing", Type: CheckTypeManual},
	},
	ActionSellExit: {
		{Key: "no_upcoming_catalyst", Label: "No upcoming positive catalyst", Type: CheckTypeManual},
		{Key: "considered_tax_impact", Label: "Considered tax impact", Type: CheckTypeManual},
	},
	ActionSellStop: {
		{Key: "accepted_loss", Label: "Accepted the loss", Type: CheckTypeManual},
		{Key: "no_recovery_signal", Label: "No recovery signal", Type: CheckTypeManual},
	},
	ActionHold: {},
}

// ThresholdsForRisk returns the threshold values for the given risk profile.
func ThresholdsForRisk(rp portfolio.RiskProfile) Thresholds {
	if t, ok := thresholdsByRisk[rp]; ok {
		return t
	}
	return thresholdsByRisk[portfolio.RiskProfileModerate]
}

// AutoCheckDefs returns the auto-check definitions for the given action type.
func AutoCheckDefs(action ActionType) []CheckDefinition {
	return autoCheckDefs[action]
}

// ManualCheckDefs returns the manual-check definitions for the given action type.
func ManualCheckDefs(action ActionType) []CheckDefinition {
	return manualCheckDefs[action]
}
