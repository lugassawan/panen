package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
	"github.com/lugassawan/panen/backend/domain/shared"
)

const (
	balanceSheetUpsert = `INSERT INTO balance_sheets
		(id, ticker, fiscal_year, quarter, period,
		 total_assets, total_current_assets, cash_and_equivalents, receivables, inventory,
		 intangible_assets, total_liabilities, total_current_liab, long_term_debt, total_debt,
		 total_equity, retained_earnings, shares_outstanding,
		 fetched_at, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ticker, fiscal_year, quarter, source) DO UPDATE SET
		period = excluded.period,
		total_assets = excluded.total_assets, total_current_assets = excluded.total_current_assets,
		cash_and_equivalents = excluded.cash_and_equivalents, receivables = excluded.receivables,
		inventory = excluded.inventory, intangible_assets = excluded.intangible_assets,
		total_liabilities = excluded.total_liabilities, total_current_liab = excluded.total_current_liab,
		long_term_debt = excluded.long_term_debt, total_debt = excluded.total_debt,
		total_equity = excluded.total_equity, retained_earnings = excluded.retained_earnings,
		shares_outstanding = excluded.shares_outstanding, fetched_at = excluded.fetched_at`
	balanceSheetGetByTicker = `SELECT id, ticker, fiscal_year, quarter, period,
		total_assets, total_current_assets, cash_and_equivalents, receivables, inventory,
		intangible_assets, total_liabilities, total_current_liab, long_term_debt, total_debt,
		total_equity, retained_earnings, shares_outstanding,
		fetched_at, source
		FROM balance_sheets
		WHERE ticker = ? AND source = ? AND period = ?
		ORDER BY fiscal_year DESC, quarter DESC`
	balanceSheetLatestFetchedAt = `SELECT MAX(fetched_at) FROM balance_sheets
		WHERE ticker = ? AND source = ?`
)

// BalanceSheetRepo implements financials.BalanceSheetRepo.
type BalanceSheetRepo struct {
	db *sql.DB
}

// NewBalanceSheetRepo creates a new BalanceSheetRepo.
func NewBalanceSheetRepo(db *sql.DB) *BalanceSheetRepo {
	return &BalanceSheetRepo{db: db}
}

func (r *BalanceSheetRepo) BulkUpsert(ctx context.Context, sheets []financials.BalanceSheet) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stmt, err := tx.PrepareContext(ctx, balanceSheetUpsert)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range sheets {
		s := &sheets[i]
		if s.ID == "" {
			s.ID = shared.NewID()
		}
		_, err := stmt.ExecContext(ctx,
			s.ID, s.Ticker, s.FiscalYear, s.Quarter, string(s.Period),
			s.TotalAssets, s.TotalCurrentAssets, s.CashAndEquivalents, s.Receivables, s.Inventory,
			s.IntangibleAssets, s.TotalLiabilities, s.TotalCurrentLiab, s.LongTermDebt, s.TotalDebt,
			s.TotalEquity, s.RetainedEarnings, s.SharesOutstanding,
			formatTime(s.FetchedAt), s.Source)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *BalanceSheetRepo) GetByTicker(
	ctx context.Context,
	ticker, source string,
	period financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	return queryAll(ctx, r.db, balanceSheetGetByTicker, scanBalanceSheet, ticker, source, string(period))
}

func (r *BalanceSheetRepo) LatestFetchedAt(
	ctx context.Context,
	ticker, source string,
) (time.Time, error) {
	var dateStr sql.NullString
	err := r.db.QueryRowContext(ctx, balanceSheetLatestFetchedAt, ticker, source).Scan(&dateStr)
	if err != nil {
		return time.Time{}, err
	}
	if !dateStr.Valid {
		return time.Time{}, nil
	}
	return parseTime(dateStr.String)
}

func scanBalanceSheet(scan func(dest ...any) error) (financials.BalanceSheet, error) {
	var s financials.BalanceSheet
	var periodStr, fetchedAtStr string
	if err := scan(
		&s.ID, &s.Ticker, &s.FiscalYear, &s.Quarter, &periodStr,
		&s.TotalAssets, &s.TotalCurrentAssets, &s.CashAndEquivalents, &s.Receivables, &s.Inventory,
		&s.IntangibleAssets, &s.TotalLiabilities, &s.TotalCurrentLiab, &s.LongTermDebt, &s.TotalDebt,
		&s.TotalEquity, &s.RetainedEarnings, &s.SharesOutstanding,
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
