package payday

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// PaydayEvent represents a scheduled monthly payday and its lifecycle state.
type PaydayEvent struct {
	ID          string
	Month       string
	PortfolioID string
	Expected    float64
	Actual      float64
	Status      Status
	DeferUntil  *time.Time
	ConfirmedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CashFlow represents a cash flow transaction within a portfolio.
type CashFlow struct {
	ID          string
	PortfolioID string
	Type        FlowType
	Amount      float64
	Date        time.Time
	Note        string
	CreatedAt   time.Time
}

// NewPaydayEvent creates a new PaydayEvent with generated ID, SCHEDULED status, and timestamps.
func NewPaydayEvent(month, portfolioID string, expected float64) *PaydayEvent {
	now := time.Now().UTC()
	return &PaydayEvent{
		ID:          shared.NewID(),
		Month:       month,
		PortfolioID: portfolioID,
		Expected:    expected,
		Status:      StatusScheduled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewCashFlow creates a new CashFlow with generated ID and timestamp.
func NewCashFlow(portfolioID string, flowType FlowType, amount float64, date time.Time, note string) *CashFlow {
	return &CashFlow{
		ID:          shared.NewID(),
		PortfolioID: portfolioID,
		Type:        flowType,
		Amount:      amount,
		Date:        date,
		Note:        note,
		CreatedAt:   time.Now().UTC(),
	}
}
