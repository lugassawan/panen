package presenter

import (
	"context"
	"errors"
	"fmt"
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
	h := &PaydayHandler{}
	h.Bind(ctx, payday)
	return h
}

func (h *PaydayHandler) Bind(ctx context.Context, payday *usecase.PaydayService) {
	h.ctx = ctx
	h.payday = payday
}

// GetPaydayDay returns the configured payday day.
func (h *PaydayHandler) GetPaydayDay() (int, error) {
	day, err := h.payday.GetPaydayDay(h.ctx)
	if err != nil {
		return 0, fmt.Errorf("get payday day: %w", err)
	}
	return day, nil
}

// SavePaydayDay persists the payday day setting.
func (h *PaydayHandler) SavePaydayDay(day int) error {
	if err := h.payday.SavePaydayDay(h.ctx, day); err != nil {
		return fmt.Errorf("save payday day: %w", err)
	}
	return nil
}

// GetCurrentMonthStatus returns the payday status for the current month.
// Returns nil with no error when payday is not configured.
func (h *PaydayHandler) GetCurrentMonthStatus() (*MonthlyPaydayResponse, error) {
	status, err := h.payday.GetCurrentMonthStatus(h.ctx)
	if err != nil {
		if errors.Is(err, usecase.ErrPaydayNotConfigured) {
			return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
		}
		return nil, fmt.Errorf("get current month status: %w", err)
	}
	return newMonthlyPaydayResponse(status), nil
}

// ConfirmPayday marks the current month's payday as confirmed for a portfolio.
func (h *PaydayHandler) ConfirmPayday(portfolioID string, actualAmount float64) error {
	if err := h.payday.ConfirmPayday(h.ctx, portfolioID, actualAmount); err != nil {
		return fmt.Errorf("confirm payday: %w", err)
	}
	return nil
}

// DeferPayday defers the current month's payday to a later date.
func (h *PaydayHandler) DeferPayday(portfolioID string, deferUntil string) error {
	t, err := time.Parse(dateLayout, deferUntil)
	if err != nil {
		return fmt.Errorf("defer payday: %w", err)
	}
	if err := h.payday.DeferPayday(h.ctx, portfolioID, t); err != nil {
		return fmt.Errorf("defer payday: %w", err)
	}
	return nil
}

// SkipPayday marks the current month's payday as skipped for a portfolio.
func (h *PaydayHandler) SkipPayday(portfolioID string) error {
	if err := h.payday.SkipPayday(h.ctx, portfolioID); err != nil {
		return fmt.Errorf("skip payday: %w", err)
	}
	return nil
}

// GetPaydayHistory returns payday statuses for all past months.
func (h *PaydayHandler) GetPaydayHistory() ([]*MonthlyPaydayResponse, error) {
	history, err := h.payday.GetPaydayHistory(h.ctx)
	if err != nil {
		return nil, fmt.Errorf("get payday history: %w", err)
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
		return nil, fmt.Errorf("get cash flow summary: %w", err)
	}
	return newCashFlowSummaryResponse(summary), nil
}
