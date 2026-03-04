package checklist

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// ChecklistResult represents the outcome of a checklist evaluation for a stock action.
type ChecklistResult struct {
	ID           string
	PortfolioID  string
	Ticker       string
	Action       ActionType
	ManualChecks map[string]bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewChecklistResult creates a new ChecklistResult with generated ID, empty manual checks, and timestamps.
func NewChecklistResult(portfolioID, ticker string, action ActionType) *ChecklistResult {
	now := time.Now().UTC()
	return &ChecklistResult{
		ID:           shared.NewID(),
		PortfolioID:  portfolioID,
		Ticker:       ticker,
		Action:       action,
		ManualChecks: map[string]bool{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
