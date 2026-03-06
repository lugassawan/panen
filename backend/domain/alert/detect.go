package alert

import (
	"math"

	"github.com/lugassawan/panen/backend/domain/stock"
)

// metricExtractor pulls a comparable value from stock.Data.
type metricExtractor struct {
	name    string
	extract func(d *stock.Data) float64
	// specialCheck overrides default severity logic when non-nil.
	specialCheck func(prev, curr float64) *Severity
}

var trackedMetrics = []metricExtractor{
	{
		name:    "roe",
		extract: func(d *stock.Data) float64 { return d.ROE },
	},
	{
		name:    "der",
		extract: func(d *stock.Data) float64 { return d.DER },
		specialCheck: func(prev, curr float64) *Severity {
			if prev < 1.0 && curr >= 1.0 {
				s := SeverityWarning
				return &s
			}
			return nil
		},
	},
	{
		name:    "eps",
		extract: func(d *stock.Data) float64 { return d.EPS },
		specialCheck: func(prev, curr float64) *Severity {
			if prev > 0 && curr <= 0 {
				s := SeverityCritical
				return &s
			}
			return nil
		},
	},
	{
		name:    "pbv",
		extract: func(d *stock.Data) float64 { return d.PBV },
	},
	{
		name:    "per",
		extract: func(d *stock.Data) float64 { return d.PER },
	},
	{
		name:    "dividend_yield",
		extract: func(d *stock.Data) float64 { return d.DividendYield },
		specialCheck: func(prev, curr float64) *Severity {
			if prev > 0 && curr == 0 {
				s := SeverityCritical
				return &s
			}
			return nil
		},
	},
	{
		name:    "payout_ratio",
		extract: func(d *stock.Data) float64 { return d.PayoutRatio },
	},
}

// DetectChanges compares two snapshots and returns alerts for significant changes.
// It is a pure function with no side effects.
func DetectChanges(prev, curr *stock.Data) []*FundamentalAlert {
	if prev == nil || curr == nil {
		return nil
	}

	var alerts []*FundamentalAlert

	for _, m := range trackedMetrics {
		oldVal := m.extract(prev)
		newVal := m.extract(curr)
		changePct := relativeChange(oldVal, newVal)

		// Check for special threshold-crossing rules first.
		if m.specialCheck != nil {
			if sev := m.specialCheck(oldVal, newVal); sev != nil {
				alerts = append(alerts, NewFundamentalAlert(
					curr.Ticker, m.name, *sev, oldVal, newVal, changePct,
				))
				continue
			}
		}

		severity := classifySeverity(changePct)
		if severity == "" {
			continue
		}

		alerts = append(alerts, NewFundamentalAlert(
			curr.Ticker, m.name, severity, oldVal, newVal, changePct,
		))
	}

	return alerts
}

// relativeChange calculates the relative percentage change between two values.
// Returns 0 if old value is 0.
func relativeChange(old, cur float64) float64 {
	if old == 0 {
		return 0
	}
	return ((cur - old) / math.Abs(old)) * 100
}

// classifySeverity maps a relative % change to a severity level.
// Returns empty string if change is below threshold.
func classifySeverity(changePct float64) Severity {
	abs := math.Abs(changePct)
	switch {
	case abs >= 30:
		return SeverityCritical
	case abs >= 15:
		return SeverityWarning
	case abs >= 5:
		return SeverityMinor
	default:
		return ""
	}
}
