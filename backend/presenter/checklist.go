package presenter

import (
	"context"

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
	return &ChecklistHandler{ctx: ctx, checklists: checklists}
}

// EvaluateChecklist evaluates a checklist for a holding.
func (h *ChecklistHandler) EvaluateChecklist(portfolioID, ticker, action string) (*ChecklistEvaluationResponse, error) {
	actionType, err := checklist.ParseActionType(action)
	if err != nil {
		return nil, err
	}
	eval, err := h.checklists.Evaluate(h.ctx, portfolioID, ticker, actionType)
	if err != nil {
		return nil, err
	}
	return newChecklistEvaluationResponse(eval), nil
}

// ToggleManualCheck toggles a manual check item.
func (h *ChecklistHandler) ToggleManualCheck(portfolioID, ticker, action, checkKey string, completed bool) error {
	actionType, err := checklist.ParseActionType(action)
	if err != nil {
		return err
	}
	return h.checklists.ToggleManualCheck(h.ctx, portfolioID, ticker, actionType, checkKey, completed)
}

// ResetChecklist resets a checklist's manual checks.
func (h *ChecklistHandler) ResetChecklist(portfolioID, ticker, action string) error {
	actionType, err := checklist.ParseActionType(action)
	if err != nil {
		return err
	}
	return h.checklists.ResetChecklist(h.ctx, portfolioID, ticker, actionType)
}

// AvailableActions returns available action types for a holding.
func (h *ChecklistHandler) AvailableActions(portfolioID, ticker string) ([]string, error) {
	actions, err := h.checklists.AvailableActions(h.ctx, portfolioID, ticker)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(actions))
	for i, a := range actions {
		result[i] = string(a)
	}
	return result, nil
}
