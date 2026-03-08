package database

import (
	"context"
	"database/sql"

	"github.com/lugassawan/panen/backend/domain/crashplaybook"
)

const (
	crashCapitalUpsert = `INSERT INTO crash_capital
		(id, portfolio_id, amount, deployed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(portfolio_id) DO UPDATE SET
		amount = excluded.amount, deployed = excluded.deployed,
		updated_at = excluded.updated_at`
	crashCapitalGetByPortfolioID = `SELECT id, portfolio_id, amount, deployed,
		created_at, updated_at FROM crash_capital WHERE portfolio_id = ?`
)

// CrashCapitalRepo implements crashplaybook.CrashCapitalRepository.
type CrashCapitalRepo struct {
	db *sql.DB
}

// NewCrashCapitalRepo creates a new CrashCapitalRepo.
func NewCrashCapitalRepo(db *sql.DB) *CrashCapitalRepo {
	return &CrashCapitalRepo{db: db}
}

func (r *CrashCapitalRepo) Upsert(ctx context.Context, cc *crashplaybook.CrashCapital) error {
	_, err := r.db.ExecContext(ctx, crashCapitalUpsert,
		cc.ID, cc.PortfolioID, cc.Amount, cc.Deployed,
		formatTime(cc.CreatedAt), formatTime(cc.UpdatedAt))
	return err
}

func (r *CrashCapitalRepo) GetByPortfolioID(
	ctx context.Context,
	portfolioID string,
) (*crashplaybook.CrashCapital, error) {
	return QueryRow(ctx, r.db, crashCapitalGetByPortfolioID, scanCrashCapital, portfolioID)
}

func scanCrashCapital(scan func(dest ...any) error) (*crashplaybook.CrashCapital, error) {
	var cc crashplaybook.CrashCapital
	var createdAt, updatedAt string
	if err := scan(&cc.ID, &cc.PortfolioID, &cc.Amount, &cc.Deployed, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	var err error
	if cc.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if cc.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &cc, nil
}
