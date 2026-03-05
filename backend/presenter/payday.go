package presenter

import (
	"context"
	"errors"
	"time"

	"github.com/lugassawan/panen/backend/usecase"
)

// PaydayHandler handles payday scheduling and cash flow requests.
type PaydayHandler struct {
	ctx    context.Context
	payday *usecase.PaydayService
}

// NewPaydayHandler creates a new PaydayHandler.
func NewPaydayHandler(ctx context.Context, payday *usecase.PaydayService) *PaydayHandler {
	return &PaydayHandler{ctx: ctx, payday: payday}
}

// GetPaydayDay returns the configured payday day.
func (h *PaydayHandler) GetPaydayDay() (int, error) {
	return h.payday.GetPaydayDay(h.ctx)
}

// SavePaydayDay persists the payday day setting.
func (h *PaydayHandler) SavePaydayDay(day int) error {
	return h.payday.SavePaydayDay(h.ctx, day)
}

// GetCurrentMonthStatus returns the payday status for the current month.
// Returns nil with no error when payday is not configured.
func (h *PaydayHandler) GetCurrentMonthStatus() (*MonthlyPaydayResponse, error) {
	status, err := h.payday.GetCurrentMonthStatus(h.ctx)
	if err != nil {
		if errors.Is(err, usecase.ErrPaydayNotConfigured) {
			return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
		}
		return nil, err
	}
	return newMonthlyPaydayResponse(status), nil
}

// ConfirmPayday marks the current month's payday as confirmed for a portfolio.
func (h *PaydayHandler) ConfirmPayday(portfolioID string, actualAmount float64) error {
	return h.payday.ConfirmPayday(h.ctx, portfolioID, actualAmount)
}

// DeferPayday defers the current month's payday to a later date.
func (h *PaydayHandler) DeferPayday(portfolioID string, deferUntil string) error {
	t, err := time.Parse(dateLayout, deferUntil)
	if err != nil {
		return err
	}
	return h.payday.DeferPayday(h.ctx, portfolioID, t)
}

// SkipPayday marks the current month's payday as skipped for a portfolio.
func (h *PaydayHandler) SkipPayday(portfolioID string) error {
	return h.payday.SkipPayday(h.ctx, portfolioID)
}

// GetPaydayHistory returns payday statuses for all past months.
func (h *PaydayHandler) GetPaydayHistory() ([]*MonthlyPaydayResponse, error) {
	history, err := h.payday.GetPaydayHistory(h.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*MonthlyPaydayResponse, len(history))
	for i, status := range history {
		result[i] = newMonthlyPaydayResponse(status)
	}
	return result, nil
}

// GetCashFlowSummary returns a cash flow summary for a portfolio.
func (h *PaydayHandler) GetCashFlowSummary(portfolioID string) (*CashFlowSummaryResponse, error) {
	summary, err := h.payday.GetCashFlowSummary(h.ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	return newCashFlowSummaryResponse(summary), nil
}
