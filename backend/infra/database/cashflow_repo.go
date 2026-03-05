package database

import (
	"context"
	"database/sql"

	"github.com/lugassawan/panen/backend/domain/payday"
)

const (
	cashFlowInsert = `INSERT INTO cash_flows
		(id, portfolio_id, type, amount, date, note, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	cashFlowListByPortfolioID = `SELECT id, portfolio_id, type, amount, date, note, created_at
		FROM cash_flows WHERE portfolio_id = ? ORDER BY date DESC`
	cashFlowDelete = `DELETE FROM cash_flows WHERE id = ?`
)

// CashFlowRepo implements payday.CashFlowRepository.
type CashFlowRepo struct {
	db *sql.DB
}

// NewCashFlowRepo creates a new CashFlowRepo.
func NewCashFlowRepo(db *sql.DB) *CashFlowRepo {
	return &CashFlowRepo{db: db}
}

func (r *CashFlowRepo) Create(ctx context.Context, cf *payday.CashFlow) error {
	_, err := r.db.ExecContext(ctx, cashFlowInsert,
		cf.ID, cf.PortfolioID, string(cf.Type), cf.Amount,
		formatTime(cf.Date), cf.Note, formatTime(cf.CreatedAt))
	return err
}

func (r *CashFlowRepo) ListByPortfolioID(ctx context.Context, portfolioID string) ([]*payday.CashFlow, error) {
	rows, err := r.db.QueryContext(ctx, cashFlowListByPortfolioID, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []*payday.CashFlow
	for rows.Next() {
		var cf payday.CashFlow
		var flowType, date, createdAt string
		if err := rows.Scan(&cf.ID, &cf.PortfolioID, &flowType, &cf.Amount,
			&date, &cf.Note, &createdAt); err != nil {
			return nil, err
		}
		cf.Type = payday.FlowType(flowType)
		if cf.Date, err = parseTime(date); err != nil {
			return nil, err
		}
		if cf.CreatedAt, err = parseTime(createdAt); err != nil {
			return nil, err
		}
		flows = append(flows, &cf)
	}
	return flows, rows.Err()
}

func (r *CashFlowRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, cashFlowDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}
