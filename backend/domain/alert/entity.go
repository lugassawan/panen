package alert

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// Severity classifies the urgency of a fundamental change alert.
type Severity string

const (
	SeverityMinor    Severity = "MINOR"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// AlertStatus tracks the lifecycle of an alert.
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "ACTIVE"
	AlertStatusAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertStatusResolved     AlertStatus = "RESOLVED"
)

// FundamentalAlert represents a detected change in a stock's fundamentals.
type FundamentalAlert struct {
	ID         string
	Ticker     string
	Metric     string // "roe", "der", "eps", "pbv", "per", "dividend_yield", "payout_ratio"
	Severity   Severity
	OldValue   float64
	NewValue   float64
	ChangePct  float64 // relative % change
	Status     AlertStatus
	DetectedAt time.Time
	ResolvedAt *time.Time
}

// NewFundamentalAlert creates a new active alert with a generated ID.
func NewFundamentalAlert(
	ticker, metric string,
	severity Severity,
	oldValue, newValue, changePct float64,
) *FundamentalAlert {
	return &FundamentalAlert{
		ID:         shared.NewID(),
		Ticker:     ticker,
		Metric:     metric,
		Severity:   severity,
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangePct:  changePct,
		Status:     AlertStatusActive,
		DetectedAt: time.Now().UTC(),
	}
}
