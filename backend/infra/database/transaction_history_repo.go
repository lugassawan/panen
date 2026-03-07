package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lugassawan/panen/backend/domain/transaction"
)

const txHistoryBase = `
SELECT id, type, date, ticker, portfolio_id, portfolio_name,
       lots, price, fee, tax, total, created_at
FROM (
    SELECT bt.id, 'BUY' AS type, bt.date, h.ticker,
           p.id AS portfolio_id, p.name AS portfolio_name,
           bt.lots, bt.price, bt.fee, 0.0 AS tax,
           (bt.price * bt.lots * 100) + bt.fee AS total,
           bt.created_at
    FROM buy_transactions bt
    JOIN holdings h ON h.id = bt.holding_id
    JOIN portfolios p ON p.id = h.portfolio_id

    UNION ALL

    SELECT st.id, 'SELL' AS type, st.date, h.ticker,
           p.id AS portfolio_id, p.name AS portfolio_name,
           st.lots, st.price, st.fee, st.tax,
           (st.price * st.lots * 100) - st.fee - st.tax AS total,
           st.created_at
    FROM sell_transactions st
    JOIN holdings h ON h.id = st.holding_id
    JOIN portfolios p ON p.id = h.portfolio_id

    UNION ALL

    SELECT cf.id, 'DIVIDEND' AS type, cf.date, '' AS ticker,
           p.id AS portfolio_id, p.name AS portfolio_name,
           0 AS lots, 0.0 AS price, 0.0 AS fee, 0.0 AS tax,
           cf.amount AS total,
           cf.created_at
    FROM cash_flows cf
    JOIN portfolios p ON p.id = cf.portfolio_id
    WHERE cf.type = 'DIVIDEND'
) AS unified`

const txSummaryBase = `
SELECT
    COALESCE(SUM(CASE WHEN type = 'BUY' THEN total ELSE 0 END), 0),
    COALESCE(SUM(CASE WHEN type = 'SELL' THEN total ELSE 0 END), 0),
    COALESCE(SUM(CASE WHEN type = 'DIVIDEND' THEN total ELSE 0 END), 0),
    COALESCE(SUM(fee), 0),
    COUNT(*)
FROM (
    SELECT 'BUY' AS type,
           bt.date, h.ticker, p.id AS portfolio_id,
           bt.fee,
           (bt.price * bt.lots * 100) + bt.fee AS total
    FROM buy_transactions bt
    JOIN holdings h ON h.id = bt.holding_id
    JOIN portfolios p ON p.id = h.portfolio_id

    UNION ALL

    SELECT 'SELL' AS type,
           st.date, h.ticker, p.id AS portfolio_id,
           st.fee,
           (st.price * st.lots * 100) - st.fee - st.tax AS total
    FROM sell_transactions st
    JOIN holdings h ON h.id = st.holding_id
    JOIN portfolios p ON p.id = h.portfolio_id

    UNION ALL

    SELECT 'DIVIDEND' AS type,
           cf.date, '' AS ticker, cf.portfolio_id,
           0.0 AS fee,
           cf.amount AS total
    FROM cash_flows cf
    WHERE cf.type = 'DIVIDEND'
) AS unified`

// sortColumns maps user-facing sort field names to SQL columns.
var sortColumns = map[string]string{
	"date":      "date",
	"type":      "type",
	"ticker":    "ticker",
	"portfolio": "portfolio_name",
	"total":     "total",
}

// TransactionHistoryRepo implements transaction.Repository.
type TransactionHistoryRepo struct {
	db *sql.DB
}

// NewTransactionHistoryRepo creates a new TransactionHistoryRepo.
func NewTransactionHistoryRepo(db *sql.DB) *TransactionHistoryRepo {
	return &TransactionHistoryRepo{db: db}
}

func (r *TransactionHistoryRepo) List(ctx context.Context, filter transaction.Filter) ([]transaction.Record, error) {
	query, args := buildHistoryQuery(txHistoryBase, filter, true)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []transaction.Record
	for rows.Next() {
		var rec transaction.Record
		var txType, date, createdAt string
		if err := rows.Scan(
			&rec.ID, &txType, &date, &rec.Ticker,
			&rec.PortfolioID, &rec.PortfolioName,
			&rec.Lots, &rec.Price, &rec.Fee, &rec.Tax,
			&rec.Total, &createdAt,
		); err != nil {
			return nil, err
		}
		rec.Type = transaction.Type(txType)
		if rec.Date, err = parseTime(date); err != nil {
			return nil, err
		}
		if rec.CreatedAt, err = parseTime(createdAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *TransactionHistoryRepo) Summarize(
	ctx context.Context,
	filter transaction.Filter,
) (*transaction.Summary, error) {
	query, args := buildHistoryQuery(txSummaryBase, filter, false)

	var s transaction.Summary
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&s.TotalBuyAmount, &s.TotalSellAmount, &s.TotalDividendAmount,
		&s.TotalFees, &s.TransactionCount,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func buildHistoryQuery(base string, f transaction.Filter, withOrder bool) (string, []any) {
	var b strings.Builder
	var args []any

	b.WriteString(base)

	var clauses []string
	if f.PortfolioID != "" {
		clauses = append(clauses, "portfolio_id = ?")
		args = append(args, f.PortfolioID)
	}
	if f.Ticker != "" {
		clauses = append(clauses, "ticker LIKE '%' || ? || '%'")
		args = append(args, f.Ticker)
	}
	if f.Type != "" {
		clauses = append(clauses, "type = ?")
		args = append(args, f.Type)
	}
	if f.DateFrom != nil {
		clauses = append(clauses, "date >= ?")
		args = append(args, formatTime(*f.DateFrom))
	}
	if f.DateTo != nil {
		clauses = append(clauses, "date <= ?")
		args = append(args, formatTime(*f.DateTo))
	}

	if len(clauses) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(strings.Join(clauses, " AND "))
	}

	if withOrder {
		writeOrderBy(&b, f)
	}

	return b.String(), args
}

func writeOrderBy(b *strings.Builder, f transaction.Filter) {
	col, ok := sortColumns[f.SortField]
	if !ok {
		b.WriteString(" ORDER BY date DESC, created_at DESC")
		return
	}

	dir := "DESC"
	if f.SortAsc {
		dir = "ASC"
	}

	fmt.Fprintf(b, " ORDER BY %s %s", col, dir)
	if col != "date" {
		b.WriteString(", date DESC")
	}
	b.WriteString(", created_at DESC")
}
