package screener

import (
	"testing"

	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

func TestCriteriaForRisk(t *testing.T) {
	tests := []struct {
		name    string
		profile valuation.RiskProfile
		minROE  float64
		maxDER  float64
	}{
		{"conservative", valuation.RiskConservative, 15, 0.8},
		{"moderate", valuation.RiskModerate, 12, 1.0},
		{"aggressive", valuation.RiskAggressive, 8, 1.5},
		{"unknown defaults to moderate", "UNKNOWN", 12, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CriteriaForRisk(tt.profile)
			if c.MinROE != tt.minROE {
				t.Errorf("MinROE = %v, want %v", c.MinROE, tt.minROE)
			}
			if c.MaxDER != tt.maxDER {
				t.Errorf("MaxDER = %v, want %v", c.MaxDER, tt.maxDER)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name       string
		data       *stock.Data
		criteria   Criteria
		val        *valuation.ValuationResult
		wantNil    bool
		wantPassed bool
		wantChecks int
	}{
		{
			name:    "nil data returns nil",
			data:    nil,
			wantNil: true,
		},
		{
			name:       "conservative pass",
			data:       &stock.Data{ROE: 20, DER: 0.5},
			criteria:   CriteriaForRisk(valuation.RiskConservative),
			val:        &valuation.ValuationResult{Verdict: valuation.VerdictUndervalued},
			wantPassed: true,
			wantChecks: 2,
		},
		{
			name:       "conservative fail ROE",
			data:       &stock.Data{ROE: 10, DER: 0.5},
			criteria:   CriteriaForRisk(valuation.RiskConservative),
			val:        &valuation.ValuationResult{Verdict: valuation.VerdictFair},
			wantPassed: false,
			wantChecks: 2,
		},
		{
			name:       "conservative fail DER",
			data:       &stock.Data{ROE: 20, DER: 1.0},
			criteria:   CriteriaForRisk(valuation.RiskConservative),
			val:        &valuation.ValuationResult{Verdict: valuation.VerdictFair},
			wantPassed: false,
			wantChecks: 2,
		},
		{
			name:       "moderate pass",
			data:       &stock.Data{ROE: 14, DER: 0.9},
			criteria:   CriteriaForRisk(valuation.RiskModerate),
			val:        &valuation.ValuationResult{Verdict: valuation.VerdictFair},
			wantPassed: true,
			wantChecks: 2,
		},
		{
			name:       "aggressive pass low ROE",
			data:       &stock.Data{ROE: 9, DER: 1.2},
			criteria:   CriteriaForRisk(valuation.RiskAggressive),
			val:        nil,
			wantPassed: true,
			wantChecks: 2,
		},
		{
			name:       "aggressive fail both",
			data:       &stock.Data{ROE: 5, DER: 2.0},
			criteria:   CriteriaForRisk(valuation.RiskAggressive),
			val:        nil,
			wantPassed: false,
			wantChecks: 2,
		},
		{
			name:       "nil valuation still evaluates",
			data:       &stock.Data{ROE: 15, DER: 0.5},
			criteria:   CriteriaForRisk(valuation.RiskConservative),
			val:        nil,
			wantPassed: true,
			wantChecks: 2,
		},
		{
			name:       "zero metrics fail conservative",
			data:       &stock.Data{ROE: 0, DER: 0},
			criteria:   CriteriaForRisk(valuation.RiskConservative),
			val:        nil,
			wantPassed: false,
			wantChecks: 2,
		},
		{
			name:       "edge case exactly at threshold passes",
			data:       &stock.Data{ROE: 15, DER: 0.8},
			criteria:   CriteriaForRisk(valuation.RiskConservative),
			val:        nil,
			wantPassed: true,
			wantChecks: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Evaluate(tt.data, tt.criteria, tt.val)
			if tt.wantNil {
				if result != nil {
					t.Fatal("expected nil result")
				}
				return
			}
			if result == nil {
				t.Fatal("unexpected nil result")
			}
			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", result.Passed, tt.wantPassed)
			}
			if len(result.Checks) != tt.wantChecks {
				t.Errorf("len(Checks) = %d, want %d", len(result.Checks), tt.wantChecks)
			}
		})
	}
}

func TestEvaluateScore(t *testing.T) {
	criteria := CriteriaForRisk(valuation.RiskModerate)

	undervalued := Evaluate(
		&stock.Data{ROE: 20, DER: 0.5},
		criteria,
		&valuation.ValuationResult{Verdict: valuation.VerdictUndervalued},
	)
	fair := Evaluate(
		&stock.Data{ROE: 20, DER: 0.5},
		criteria,
		&valuation.ValuationResult{Verdict: valuation.VerdictFair},
	)
	overvalued := Evaluate(
		&stock.Data{ROE: 20, DER: 0.5},
		criteria,
		&valuation.ValuationResult{Verdict: valuation.VerdictOvervalued},
	)

	if undervalued.Score <= fair.Score {
		t.Errorf("undervalued score (%f) should be > fair score (%f)", undervalued.Score, fair.Score)
	}
	if fair.Score <= overvalued.Score {
		t.Errorf("fair score (%f) should be > overvalued score (%f)", fair.Score, overvalued.Score)
	}
}

func TestEvaluateCheckStatuses(t *testing.T) {
	result := Evaluate(
		&stock.Data{ROE: 10, DER: 1.5},
		CriteriaForRisk(valuation.RiskConservative),
		nil,
	)

	for _, c := range result.Checks {
		switch c.Key {
		case "roe_above_min":
			if c.Status != StatusFail {
				t.Errorf("ROE check: got %s, want FAIL", c.Status)
			}
			if c.Value != 10 {
				t.Errorf("ROE value = %f, want 10", c.Value)
			}
			if c.Limit != 15 {
				t.Errorf("ROE limit = %f, want 15", c.Limit)
			}
		case "der_below_max":
			if c.Status != StatusFail {
				t.Errorf("DER check: got %s, want FAIL", c.Status)
			}
			if c.Value != 1.5 {
				t.Errorf("DER value = %f, want 1.5", c.Value)
			}
			if c.Limit != 0.8 {
				t.Errorf("DER limit = %f, want 0.8", c.Limit)
			}
		}
	}
}
