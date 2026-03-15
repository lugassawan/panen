package financials

import (
	"context"
	"time"
)

// IncomeStatementRepo persists income statement data.
type IncomeStatementRepo interface {
	BulkUpsert(ctx context.Context, stmts []IncomeStatement) error
	GetByTicker(ctx context.Context, ticker, source string, period PeriodType) ([]IncomeStatement, error)
	LatestFetchedAt(ctx context.Context, ticker, source string) (time.Time, error)
}

// BalanceSheetRepo persists balance sheet data.
type BalanceSheetRepo interface {
	BulkUpsert(ctx context.Context, sheets []BalanceSheet) error
	GetByTicker(ctx context.Context, ticker, source string, period PeriodType) ([]BalanceSheet, error)
	LatestFetchedAt(ctx context.Context, ticker, source string) (time.Time, error)
}

// CashFlowStatementRepo persists cash flow statement data.
type CashFlowStatementRepo interface {
	BulkUpsert(ctx context.Context, stmts []CashFlowStatement) error
	GetByTicker(ctx context.Context, ticker, source string, period PeriodType) ([]CashFlowStatement, error)
	LatestFetchedAt(ctx context.Context, ticker, source string) (time.Time, error)
}
