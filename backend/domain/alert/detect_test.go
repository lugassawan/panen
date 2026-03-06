package alert

import (
	"testing"

	"github.com/lugassawan/panen/backend/domain/stock"
)

func baseData() *stock.Data {
	return &stock.Data{
		Ticker:        "BBCA",
		Price:         8000,
		EPS:           500,
		BVPS:          2500,
		ROE:           20,
		DER:           0.5,
		PBV:           3.2,
		PER:           16,
		DividendYield: 3.0,
		PayoutRatio:   40,
		Source:        "yahoo",
	}
}

func TestDetectChanges(t *testing.T) {
	tests := []struct {
		name       string
		modify     func(d *stock.Data)
		wantMetric string
		wantSev    Severity
	}{
		{
			name:       "ROE minor drop (10%)",
			modify:     func(d *stock.Data) { d.ROE = 18 },
			wantMetric: "roe",
			wantSev:    SeverityMinor,
		},
		{
			name:       "ROE warning drop (20%)",
			modify:     func(d *stock.Data) { d.ROE = 16 },
			wantMetric: "roe",
			wantSev:    SeverityWarning,
		},
		{
			name:       "ROE critical drop (35%)",
			modify:     func(d *stock.Data) { d.ROE = 13 },
			wantMetric: "roe",
			wantSev:    SeverityCritical,
		},
		{
			name:       "DER crosses above 1.0",
			modify:     func(d *stock.Data) { d.DER = 1.1 },
			wantMetric: "der",
			wantSev:    SeverityWarning,
		},
		{
			name:       "EPS goes negative",
			modify:     func(d *stock.Data) { d.EPS = -10 },
			wantMetric: "eps",
			wantSev:    SeverityCritical,
		},
		{
			name:       "EPS goes to zero",
			modify:     func(d *stock.Data) { d.EPS = 0 },
			wantMetric: "eps",
			wantSev:    SeverityCritical,
		},
		{
			name:       "DividendYield drops to zero",
			modify:     func(d *stock.Data) { d.DividendYield = 0 },
			wantMetric: "dividend_yield",
			wantSev:    SeverityCritical,
		},
		{
			name:       "PBV minor change (8%)",
			modify:     func(d *stock.Data) { d.PBV = 3.456 },
			wantMetric: "pbv",
			wantSev:    SeverityMinor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev := baseData()
			curr := baseData()
			tt.modify(curr)

			alerts := DetectChanges(prev, curr)

			var found *FundamentalAlert
			for _, a := range alerts {
				if a.Metric == tt.wantMetric {
					found = a
					break
				}
			}
			if found == nil {
				t.Fatalf("expected alert for metric %q, got none (total alerts: %d)", tt.wantMetric, len(alerts))
			}
			if found.Severity != tt.wantSev {
				t.Errorf("severity = %q, want %q", found.Severity, tt.wantSev)
			}
			if found.Ticker != "BBCA" {
				t.Errorf("ticker = %q, want BBCA", found.Ticker)
			}
		})
	}
}

func TestDetectChangesNoChange(t *testing.T) {
	prev := baseData()
	curr := baseData()

	alerts := DetectChanges(prev, curr)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts for identical data, got %d", len(alerts))
	}
}

func TestDetectChangesBelowThreshold(t *testing.T) {
	prev := baseData()
	curr := baseData()
	curr.ROE = 19.5 // 2.5% change — below 5% threshold

	alerts := DetectChanges(prev, curr)
	for _, a := range alerts {
		if a.Metric == "roe" {
			t.Errorf("expected no alert for 2.5%% ROE change, got severity %q", a.Severity)
		}
	}
}

func TestDetectChangesNilInputs(t *testing.T) {
	prev := baseData()

	if alerts := DetectChanges(nil, prev); alerts != nil {
		t.Error("expected nil for nil prev")
	}
	if alerts := DetectChanges(prev, nil); alerts != nil {
		t.Error("expected nil for nil curr")
	}
	if alerts := DetectChanges(nil, nil); alerts != nil {
		t.Error("expected nil for both nil")
	}
}

func TestDetectChangesZeroOldValue(t *testing.T) {
	prev := baseData()
	prev.DividendYield = 0
	curr := baseData()
	curr.DividendYield = 3.0

	alerts := DetectChanges(prev, curr)
	for _, a := range alerts {
		if a.Metric == "dividend_yield" {
			t.Errorf("expected no alert when old value is 0, got severity %q", a.Severity)
		}
	}
}

func TestRelativeChange(t *testing.T) {
	tests := []struct {
		name     string
		old, new float64
		want     float64
	}{
		{"positive drop", 20, 14, -30},
		{"positive rise", 20, 26, 30},
		{"zero old", 0, 10, 0},
		{"no change", 15, 15, 0},
		{"negative to more negative", -10, -13, -30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := relativeChange(tt.old, tt.new)
			if diff := got - tt.want; diff > 0.01 || diff < -0.01 {
				t.Errorf("relativeChange(%v, %v) = %v, want %v", tt.old, tt.new, got, tt.want)
			}
		})
	}
}

func TestClassifySeverity(t *testing.T) {
	tests := []struct {
		changePct float64
		want      Severity
	}{
		{3, ""},
		{-3, ""},
		{5, SeverityMinor},
		{-10, SeverityMinor},
		{15, SeverityWarning},
		{-25, SeverityWarning},
		{30, SeverityCritical},
		{-50, SeverityCritical},
	}

	for _, tt := range tests {
		got := classifySeverity(tt.changePct)
		if got != tt.want {
			t.Errorf("classifySeverity(%v) = %q, want %q", tt.changePct, got, tt.want)
		}
	}
}

func TestDetectChangesDERCrossingFromAbove(t *testing.T) {
	// DER going from 1.05 to 0.99 is small change (~5.7%) — should be MINOR, not WARNING.
	// The special WARNING rule only fires when crossing from below 1.0 to above.
	prev := baseData()
	prev.DER = 1.05
	curr := baseData()
	curr.DER = 0.99

	alerts := DetectChanges(prev, curr)
	for _, a := range alerts {
		if a.Metric == "der" && a.Severity == SeverityWarning {
			t.Errorf(
				"DER crossing from above 1.0 to below should not trigger WARNING special rule, got severity %q",
				a.Severity,
			)
		}
	}
}
