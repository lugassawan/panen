package checklist

import (
	"testing"

	"github.com/lugassawan/panen/backend/domain/portfolio"
)

func TestThresholdsForRisk(t *testing.T) {
	tests := []struct {
		name string
		risk portfolio.RiskProfile
		want Thresholds
	}{
		{
			name: "conservative",
			risk: portfolio.RiskProfileConservative,
			want: Thresholds{
				MinROE:         15,
				MaxDER:         0.8,
				MaxPositionPct: 10,
				MinDY:          5,
				MaxPayoutRatio: 60,
			},
		},
		{
			name: "moderate",
			risk: portfolio.RiskProfileModerate,
			want: Thresholds{
				MinROE:         12,
				MaxDER:         1.0,
				MaxPositionPct: 20,
				MinDY:          3,
				MaxPayoutRatio: 75,
			},
		},
		{
			name: "aggressive",
			risk: portfolio.RiskProfileAggressive,
			want: Thresholds{
				MinROE:         8,
				MaxDER:         1.5,
				MaxPositionPct: 35,
				MinDY:          2,
				MaxPayoutRatio: 90,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ThresholdsForRisk(tt.risk)
			if got != tt.want {
				t.Errorf("ThresholdsForRisk(%q) = %+v, want %+v", tt.risk, got, tt.want)
			}
		})
	}
}

func TestThresholdsForRiskUnknown(t *testing.T) {
	got := ThresholdsForRisk("UNKNOWN")
	want := ThresholdsForRisk(portfolio.RiskProfileModerate)
	if got != want {
		t.Errorf("ThresholdsForRisk(UNKNOWN) = %+v, want moderate defaults %+v", got, want)
	}
}

func TestAutoCheckDefs(t *testing.T) {
	tests := []struct {
		action ActionType
		count  int
	}{
		{action: ActionBuy, count: 4},
		{action: ActionAverageDown, count: 5},
		{action: ActionAverageUp, count: 4},
		{action: ActionSellExit, count: 2},
		{action: ActionSellStop, count: 2},
		{action: ActionHold, count: 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			defs := AutoCheckDefs(tt.action)
			if len(defs) != tt.count {
				t.Errorf("AutoCheckDefs(%q) returned %d defs, want %d", tt.action, len(defs), tt.count)
			}
			assertDefsValid(t, defs, CheckTypeAuto)
		})
	}
}

func TestManualCheckDefs(t *testing.T) {
	tests := []struct {
		action ActionType
		count  int
	}{
		{action: ActionBuy, count: 3},
		{action: ActionAverageDown, count: 2},
		{action: ActionAverageUp, count: 2},
		{action: ActionSellExit, count: 2},
		{action: ActionSellStop, count: 2},
		{action: ActionHold, count: 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			defs := ManualCheckDefs(tt.action)
			if len(defs) != tt.count {
				t.Errorf("ManualCheckDefs(%q) returned %d defs, want %d", tt.action, len(defs), tt.count)
			}
			assertDefsValid(t, defs, CheckTypeManual)
		})
	}
}

func assertDefsValid(t *testing.T, defs []CheckDefinition, wantType CheckType) {
	t.Helper()
	for _, d := range defs {
		if d.Key == "" {
			t.Error("check definition has empty Key")
		}
		if d.Label == "" {
			t.Errorf("check definition %q has empty Label", d.Key)
		}
		if d.Type != wantType {
			t.Errorf("check definition %q has Type %q, want %q", d.Key, d.Type, wantType)
		}
	}
}
