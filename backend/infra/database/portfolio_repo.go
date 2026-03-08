package database

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lugassawan/panen/backend/domain/portfolio"
)

const (
	portfolioInsert = `INSERT INTO portfolios
		(id, brokerage_acct_id, name, mode, risk_profile, capital,
		 monthly_addition, max_stocks, universe, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	portfolioGetByID = `SELECT id, brokerage_acct_id, name, mode, risk_profile, capital,
		monthly_addition, max_stocks, universe, created_at, updated_at
		FROM portfolios WHERE id = ?`
	portfolioListAll = `SELECT id, brokerage_acct_id, name, mode, risk_profile,
		capital, monthly_addition, max_stocks, universe, created_at, updated_at
		FROM portfolios ORDER BY created_at`
	portfolioListByBrokerageAccountID = `SELECT id, brokerage_acct_id, name, mode, risk_profile,
		capital, monthly_addition, max_stocks, universe, created_at, updated_at
		FROM portfolios WHERE brokerage_acct_id = ? ORDER BY created_at`
	portfolioUpdate = `UPDATE portfolios SET name = ?, mode = ?, risk_profile = ?, capital = ?,
		monthly_addition = ?, max_stocks = ?, universe = ?, updated_at = ? WHERE id = ?`
	portfolioDelete = `DELETE FROM portfolios WHERE id = ?`
)

// PortfolioRepo implements portfolio.Repository.
type PortfolioRepo struct {
	db *sql.DB
}

// NewPortfolioRepo creates a new PortfolioRepo.
func NewPortfolioRepo(db *sql.DB) *PortfolioRepo {
	return &PortfolioRepo{db: db}
}

func (r *PortfolioRepo) Create(ctx context.Context, p *portfolio.Portfolio) error {
	universe, err := json.Marshal(p.Universe)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, portfolioInsert,
		p.ID, p.BrokerageAccountID, p.Name, string(p.Mode), string(p.RiskProfile),
		p.Capital, p.MonthlyAddition, p.MaxStocks, string(universe),
		formatTime(p.CreatedAt), formatTime(p.UpdatedAt))
	return err
}

func (r *PortfolioRepo) GetByID(ctx context.Context, id string) (*portfolio.Portfolio, error) {
	return QueryRow(ctx, r.db, portfolioGetByID, scanPortfolio, id)
}

func (r *PortfolioRepo) ListAll(ctx context.Context) ([]*portfolio.Portfolio, error) {
	return QueryAll(ctx, r.db, portfolioListAll, scanPortfolio)
}

func (r *PortfolioRepo) ListByBrokerageAccountID(
	ctx context.Context, brokerageAccountID string,
) ([]*portfolio.Portfolio, error) {
	return QueryAll(ctx, r.db, portfolioListByBrokerageAccountID, scanPortfolio, brokerageAccountID)
}

func (r *PortfolioRepo) Update(ctx context.Context, p *portfolio.Portfolio) error {
	universe, err := json.Marshal(p.Universe)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, portfolioUpdate,
		p.Name, string(p.Mode), string(p.RiskProfile), p.Capital,
		p.MonthlyAddition, p.MaxStocks, string(universe),
		formatTime(p.UpdatedAt), p.ID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *PortfolioRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, portfolioDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func scanPortfolio(scan func(dest ...any) error) (*portfolio.Portfolio, error) {
	var p portfolio.Portfolio
	var mode, riskProfile, universe, createdAt, updatedAt string
	if err := scan(
		&p.ID, &p.BrokerageAccountID, &p.Name, &mode, &riskProfile,
		&p.Capital, &p.MonthlyAddition, &p.MaxStocks, &universe,
		&createdAt, &updatedAt); err != nil {
		return nil, err
	}
	p.Mode = portfolio.Mode(mode)
	p.RiskProfile = portfolio.RiskProfile(riskProfile)
	if err := json.Unmarshal([]byte(universe), &p.Universe); err != nil {
		return nil, err
	}
	if p.Universe == nil {
		p.Universe = []string{}
	}
	var err error
	if p.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if p.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}
