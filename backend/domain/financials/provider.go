package financials

import "context"

// FinancialStatementProvider defines operations for fetching full financial
// statement data from an external source.
type FinancialStatementProvider interface {
	// Source returns the provider identifier (e.g. "fmp", "sectors").
	Source() string
	// FetchIncomeStatements returns income statements for a ticker and period type.
	FetchIncomeStatements(ctx context.Context, ticker string, period PeriodType) ([]IncomeStatement, error)
	// FetchBalanceSheets returns balance sheets for a ticker and period type.
	FetchBalanceSheets(ctx context.Context, ticker string, period PeriodType) ([]BalanceSheet, error)
	// FetchCashFlowStatements returns cash flow statements for a ticker and period type.
	FetchCashFlowStatements(ctx context.Context, ticker string, period PeriodType) ([]CashFlowStatement, error)
}
