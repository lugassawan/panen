package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/shared"
)

const (
	checklistResultUpsert = `INSERT INTO checklist_results
		(id, portfolio_id, ticker, action, manual_checks, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(portfolio_id, ticker, action)
		DO UPDATE SET manual_checks = ?, updated_at = ?`
	checklistResultGet = `SELECT id, portfolio_id, ticker, action, manual_checks,
		created_at, updated_at
		FROM checklist_results WHERE portfolio_id = ? AND ticker = ? AND action = ?`
	checklistResultDelete              = `DELETE FROM checklist_results WHERE id = ?`
	checklistResultDeleteByPortfolioID = `DELETE FROM checklist_results WHERE portfolio_id = ?`
)

// ChecklistResultRepo implements checklist.Repository.
type ChecklistResultRepo struct {
	db *sql.DB
}

// NewChecklistResultRepo creates a new ChecklistResultRepo.
func NewChecklistResultRepo(db *sql.DB) *ChecklistResultRepo {
	return &ChecklistResultRepo{db: db}
}

func (r *ChecklistResultRepo) Upsert(ctx context.Context, cr *checklist.ChecklistResult) error {
	manualChecks, err := json.Marshal(cr.ManualChecks)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, checklistResultUpsert,
		cr.ID, cr.PortfolioID, cr.Ticker, string(cr.Action),
		string(manualChecks), formatTime(cr.CreatedAt), formatTime(cr.UpdatedAt),
		string(manualChecks), formatTime(cr.UpdatedAt))
	return err
}

func (r *ChecklistResultRepo) Get(
	ctx context.Context, portfolioID, ticker string, action checklist.ActionType,
) (*checklist.ChecklistResult, error) {
	var cr checklist.ChecklistResult
	var actionStr, manualChecks, createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, checklistResultGet,
		portfolioID, ticker, string(action)).Scan(
		&cr.ID, &cr.PortfolioID, &cr.Ticker, &actionStr, &manualChecks,
		&createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	cr.Action = checklist.ActionType(actionStr)
	if err := json.Unmarshal([]byte(manualChecks), &cr.ManualChecks); err != nil {
		return nil, err
	}
	if cr.ManualChecks == nil {
		cr.ManualChecks = map[string]bool{}
	}
	if cr.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if cr.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &cr, nil
}

func (r *ChecklistResultRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, checklistResultDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *ChecklistResultRepo) DeleteByPortfolioID(ctx context.Context, portfolioID string) error {
	_, err := r.db.ExecContext(ctx, checklistResultDeleteByPortfolioID, portfolioID)
	return err
}
