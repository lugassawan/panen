package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
	"github.com/lugassawan/panen/backend/domain/shared"
)

const (
	cashFlowStmtUpsert = `INSERT INTO cash_flow_statements
		(id, ticker, fiscal_year, quarter, period,
		 operating_cash_flow, capital_expenditure, free_cash_flow, dividends_paid, net_borrowings,
		 investing_cash_flow, financing_cash_flow, net_change_in_cash,
		 fetched_at, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ticker, fiscal_year, quarter, source) DO UPDATE SET
		period = excluded.period,
		operating_cash_flow = excluded.operating_cash_flow, capital_expenditure = excluded.capital_expenditure,
		free_cash_flow = excluded.free_cash_flow, dividends_paid = excluded.dividends_paid,
		net_borrowings = excluded.net_borrowings, investing_cash_flow = excluded.investing_cash_flow,
		financing_cash_flow = excluded.financing_cash_flow, net_change_in_cash = excluded.net_change_in_cash,
		fetched_at = excluded.fetched_at`
	cashFlowStmtGetByTicker = `SELECT id, ticker, fiscal_year, quarter, period,
		operating_cash_flow, capital_expenditure, free_cash_flow, dividends_paid, net_borrowings,
		investing_cash_flow, financing_cash_flow, net_change_in_cash,
		fetched_at, source
		FROM cash_flow_statements
		WHERE ticker = ? AND source = ? AND period = ?
		ORDER BY fiscal_year DESC, quarter DESC`
	cashFlowStmtLatestFetchedAt = `SELECT MAX(fetched_at) FROM cash_flow_statements
		WHERE ticker = ? AND source = ?`
)

// CashFlowStatementRepo implements financials.CashFlowStatementRepo.
type CashFlowStatementRepo struct {
	db *sql.DB
}

// NewCashFlowStatementRepo creates a new CashFlowStatementRepo.
func NewCashFlowStatementRepo(db *sql.DB) *CashFlowStatementRepo {
	return &CashFlowStatementRepo{db: db}
}

func (r *CashFlowStatementRepo) BulkUpsert(ctx context.Context, stmts []financials.CashFlowStatement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stmt, err := tx.PrepareContext(ctx, cashFlowStmtUpsert)
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
			s.OperatingCashFlow, s.CapitalExpenditure, s.FreeCashFlow, s.DividendsPaid, s.NetBorrowings,
			s.InvestingCashFlow, s.FinancingCashFlow, s.NetChangeInCash,
			formatTime(s.FetchedAt), s.Source)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *CashFlowStatementRepo) GetByTicker(
	ctx context.Context,
	ticker, source string,
	period financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	return queryAll(ctx, r.db, cashFlowStmtGetByTicker, scanCashFlowStatement, ticker, source, string(period))
}

func (r *CashFlowStatementRepo) LatestFetchedAt(
	ctx context.Context,
	ticker, source string,
) (time.Time, error) {
	var dateStr sql.NullString
	err := r.db.QueryRowContext(ctx, cashFlowStmtLatestFetchedAt, ticker, source).Scan(&dateStr)
	if err != nil {
		return time.Time{}, err
	}
	if !dateStr.Valid {
		return time.Time{}, nil
	}
	return parseTime(dateStr.String)
}

func scanCashFlowStatement(scan func(dest ...any) error) (financials.CashFlowStatement, error) {
	var s financials.CashFlowStatement
	var periodStr, fetchedAtStr string
	if err := scan(
		&s.ID, &s.Ticker, &s.FiscalYear, &s.Quarter, &periodStr,
		&s.OperatingCashFlow, &s.CapitalExpenditure, &s.FreeCashFlow, &s.DividendsPaid, &s.NetBorrowings,
		&s.InvestingCashFlow, &s.FinancingCashFlow, &s.NetChangeInCash,
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
