package screener

import (
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

// CheckStatus indicates whether a screening check passed or failed.
type CheckStatus string

const (
	StatusPass CheckStatus = "PASS"
	StatusFail CheckStatus = "FAIL"
)

// Check holds the result of a single screening criterion.
type Check struct {
	Key    string
	Label  string
	Status CheckStatus
	Value  float64
	Limit  float64
}

// Result holds the overall screening outcome for a stock.
type Result struct {
	Passed bool
	Checks []Check
	Score  float64
}

// Evaluate screens a stock against the given criteria and valuation result.
// It returns nil if data is nil.
func Evaluate(data *stock.Data, criteria Criteria, val *valuation.ValuationResult) *Result {
	if data == nil {
		return nil
	}

	roeStatus := StatusPass
	if data.ROE < criteria.MinROE {
		roeStatus = StatusFail
	}
	derStatus := StatusPass
	if data.DER > criteria.MaxDER {
		derStatus = StatusFail
	}

	checks := []Check{
		{
			Key:    "roe_above_min",
			Label:  "ROE above minimum",
			Status: roeStatus,
			Value:  data.ROE,
			Limit:  criteria.MinROE,
		},
		{
			Key:    "der_below_max",
			Label:  "DER below maximum",
			Status: derStatus,
			Value:  data.DER,
			Limit:  criteria.MaxDER,
		},
	}

	passed := roeStatus == StatusPass && derStatus == StatusPass

	score := computeScore(data, criteria, val)

	return &Result{
		Passed: passed,
		Checks: checks,
		Score:  score,
	}
}

// computeScore produces a composite attractiveness score.
// Higher is better: ROE headroom + DER margin + verdict bonus.
func computeScore(data *stock.Data, criteria Criteria, val *valuation.ValuationResult) float64 {
	var score float64

	// ROE headroom: how far above the minimum threshold
	if criteria.MinROE > 0 {
		score += (data.ROE - criteria.MinROE) / criteria.MinROE
	}

	// DER margin: how far below the maximum threshold (inverted so lower DER = higher score)
	if criteria.MaxDER > 0 {
		score += (criteria.MaxDER - data.DER) / criteria.MaxDER
	}

	// Verdict bonus
	if val != nil {
		switch val.Verdict {
		case valuation.VerdictUndervalued:
			score += 2
		case valuation.VerdictFair:
			score += 1
		}
	}

	return score
}
