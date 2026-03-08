package presenter

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/usecase"
)

// ChecklistHandler handles checklist evaluation requests.
type ChecklistHandler struct {
	ctx        context.Context
	checklists *usecase.ChecklistService
}

// NewChecklistHandler creates a new ChecklistHandler.
func NewChecklistHandler(ctx context.Context, checklists *usecase.ChecklistService) *ChecklistHandler {
	h := &ChecklistHandler{}
	h.Bind(ctx, checklists)
	return h
}

func (h *ChecklistHandler) Bind(ctx context.Context, checklists *usecase.ChecklistService) {
	h.ctx = ctx
	h.checklists = checklists
}

// EvaluateChecklist evaluates a checklist for a holding.
func (h *ChecklistHandler) EvaluateChecklist(portfolioID, ticker, action string) (*ChecklistEvaluationResponse, error) {
	actionType, err := checklist.ParseActionType(action)
	if err != nil {
		return nil, fmt.Errorf("evaluate checklist: %w", err)
	}
	eval, err := h.checklists.Evaluate(h.ctx, portfolioID, ticker, actionType)
	if err != nil {
		return nil, fmt.Errorf("evaluate checklist: %w", err)
	}
	return newChecklistEvaluationResponse(eval), nil
}

// ToggleManualCheck toggles a manual check item.
func (h *ChecklistHandler) ToggleManualCheck(portfolioID, ticker, action, checkKey string, completed bool) error {
	actionType, err := checklist.ParseActionType(action)
	if err != nil {
		return fmt.Errorf("toggle manual check: %w", err)
	}
	if err := h.checklists.ToggleManualCheck(h.ctx, portfolioID, ticker, actionType, checkKey, completed); err != nil {
		return fmt.Errorf("toggle manual check: %w", err)
	}
	return nil
}

// ResetChecklist resets a checklist's manual checks.
func (h *ChecklistHandler) ResetChecklist(portfolioID, ticker, action string) error {
	actionType, err := checklist.ParseActionType(action)
	if err != nil {
		return fmt.Errorf("reset checklist: %w", err)
	}
	if err := h.checklists.ResetChecklist(h.ctx, portfolioID, ticker, actionType); err != nil {
		return fmt.Errorf("reset checklist: %w", err)
	}
	return nil
}

// AvailableActions returns available action types for a holding.
func (h *ChecklistHandler) AvailableActions(portfolioID, ticker string) ([]string, error) {
	actions, err := h.checklists.AvailableActions(h.ctx, portfolioID, ticker)
	if err != nil {
		return nil, fmt.Errorf("available actions: %w", err)
	}
	result := make([]string, len(actions))
	for i, a := range actions {
		result[i] = string(a)
	}
	return result, nil
}
