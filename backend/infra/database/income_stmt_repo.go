package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
	"github.com/lugassawan/panen/backend/domain/shared"
)

const (
	incomeStmtUpsert = `INSERT INTO income_statements
		(id, ticker, fiscal_year, quarter, period,
		 revenue, cost_of_revenue, gross_profit, operating_expenses, operating_income,
		 net_income, eps, eps_diluted, ebitda, interest_expense, income_tax_expense,
		 fetched_at, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ticker, fiscal_year, quarter, source) DO UPDATE SET
		period = excluded.period,
		revenue = excluded.revenue, cost_of_revenue = excluded.cost_of_revenue,
		gross_profit = excluded.gross_profit, operating_expenses = excluded.operating_expenses,
		operating_income = excluded.operating_income, net_income = excluded.net_income,
		eps = excluded.eps, eps_diluted = excluded.eps_diluted,
		ebitda = excluded.ebitda, interest_expense = excluded.interest_expense,
		income_tax_expense = excluded.income_tax_expense, fetched_at = excluded.fetched_at`
	incomeStmtGetByTicker = `SELECT id, ticker, fiscal_year, quarter, period,
		revenue, cost_of_revenue, gross_profit, operating_expenses, operating_income,
		net_income, eps, eps_diluted, ebitda, interest_expense, income_tax_expense,
		fetched_at, source
		FROM income_statements
		WHERE ticker = ? AND source = ? AND period = ?
		ORDER BY fiscal_year DESC, quarter DESC`
	incomeStmtLatestFetchedAt = `SELECT MAX(fetched_at) FROM income_statements
		WHERE ticker = ? AND source = ?`
)

// IncomeStatementRepo implements financials.IncomeStatementRepo.
type IncomeStatementRepo struct {
	db *sql.DB
}

// NewIncomeStatementRepo creates a new IncomeStatementRepo.
func NewIncomeStatementRepo(db *sql.DB) *IncomeStatementRepo {
	return &IncomeStatementRepo{db: db}
}

func (r *IncomeStatementRepo) BulkUpsert(ctx context.Context, stmts []financials.IncomeStatement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stmt, err := tx.PrepareContext(ctx, incomeStmtUpsert)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range stmts {
		s := &stmts[i]
		if s.ID == "" {
			s.ID = shared.NewID()
		}
		_, err := stmt.ExecContext(ctx,
			s.ID, s.Ticker, s.FiscalYear, s.Quarter, string(s.Period),
			s.Revenue, s.CostOfRevenue, s.GrossProfit, s.OperatingExpenses, s.OperatingIncome,
			s.NetIncome, s.EPS, s.EPSDiluted, s.EBITDA, s.InterestExpense, s.IncomeTaxExpense,
			formatTime(s.FetchedAt), s.Source)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *IncomeStatementRepo) GetByTicker(
	ctx context.Context,
	ticker, source string,
	period financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	return queryAll(ctx, r.db, incomeStmtGetByTicker, scanIncomeStatement, ticker, source, string(period))
}

func (r *IncomeStatementRepo) LatestFetchedAt(
	ctx context.Context,
	ticker, source string,
) (time.Time, error) {
	var dateStr sql.NullString
	err := r.db.QueryRowContext(ctx, incomeStmtLatestFetchedAt, ticker, source).Scan(&dateStr)
	if err != nil {
		return time.Time{}, err
	}
	if !dateStr.Valid {
		return time.Time{}, nil
	}
	return parseTime(dateStr.String)
}

func scanIncomeStatement(scan func(dest ...any) error) (financials.IncomeStatement, error) {
	var s financials.IncomeStatement
	var periodStr, fetchedAtStr string
	if err := scan(
		&s.ID, &s.Ticker, &s.FiscalYear, &s.Quarter, &periodStr,
		&s.Revenue, &s.CostOfRevenue, &s.GrossProfit, &s.OperatingExpenses, &s.OperatingIncome,
		&s.NetIncome, &s.EPS, &s.EPSDiluted, &s.EBITDA, &s.InterestExpense, &s.IncomeTaxExpense,
		&fetchedAtStr, &s.Source,
	); err != nil {
		return s, err
	}
	s.Period = financials.PeriodType(periodStr)
	var err error
	if s.FetchedAt, err = parseTime(fetchedAtStr); err != nil {
		return s, err
	}
	return s, nil
}
