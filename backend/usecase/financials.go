package usecase

import (
	"context"
	"time"

	"github.com/lugassawan/panen/backend/domain/financials"
)

const financialsCacheTTL = 24 * time.Hour

// FinancialStatementsService handles on-demand fetching and caching of financial statements.
type FinancialStatementsService struct {
	incomeRepo   financials.IncomeStatementRepo
	balanceRepo  financials.BalanceSheetRepo
	cashFlowRepo financials.CashFlowStatementRepo
	provider     financials.FinancialStatementProvider
}

// NewFinancialStatementsService creates a new FinancialStatementsService.
func NewFinancialStatementsService(
	incomeRepo financials.IncomeStatementRepo,
	balanceRepo financials.BalanceSheetRepo,
	cashFlowRepo financials.CashFlowStatementRepo,
	provider financials.FinancialStatementProvider,
) *FinancialStatementsService {
	return &FinancialStatementsService{
		incomeRepo:   incomeRepo,
		balanceRepo:  balanceRepo,
		cashFlowRepo: cashFlowRepo,
		provider:     provider,
	}
}

// GetIncomeStatements returns cached income statements, refreshing from the provider if stale.
func (s *FinancialStatementsService) GetIncomeStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.IncomeStatement, error) {
	latest, err := s.incomeRepo.LatestFetchedAt(ctx, ticker, s.provider.Source())
	if err != nil {
		return nil, err
	}

	if !isFinancialsFresh(latest) {
		stmts, err := s.provider.FetchIncomeStatements(ctx, ticker, period)
		if err != nil {
			return nil, err
		}
		if err := s.incomeRepo.BulkUpsert(ctx, stmts); err != nil {
			return nil, err
		}
	}

	return s.incomeRepo.GetByTicker(ctx, ticker, s.provider.Source(), period)
}

// GetBalanceSheets returns cached balance sheets, refreshing from the provider if stale.
func (s *FinancialStatementsService) GetBalanceSheets(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.BalanceSheet, error) {
	latest, err := s.balanceRepo.LatestFetchedAt(ctx, ticker, s.provider.Source())
	if err != nil {
		return nil, err
	}

	if !isFinancialsFresh(latest) {
		sheets, err := s.provider.FetchBalanceSheets(ctx, ticker, period)
		if err != nil {
			return nil, err
		}
		if err := s.balanceRepo.BulkUpsert(ctx, sheets); err != nil {
			return nil, err
		}
	}

	return s.balanceRepo.GetByTicker(ctx, ticker, s.provider.Source(), period)
}

// GetCashFlowStatements returns cached cash flow statements, refreshing from the provider if stale.
func (s *FinancialStatementsService) GetCashFlowStatements(
	ctx context.Context,
	ticker string,
	period financials.PeriodType,
) ([]financials.CashFlowStatement, error) {
	latest, err := s.cashFlowRepo.LatestFetchedAt(ctx, ticker, s.provider.Source())
	if err != nil {
		return nil, err
	}

	if !isFinancialsFresh(latest) {
		stmts, err := s.provider.FetchCashFlowStatements(ctx, ticker, period)
		if err != nil {
			return nil, err
		}
		if err := s.cashFlowRepo.BulkUpsert(ctx, stmts); err != nil {
			return nil, err
		}
	}

	return s.cashFlowRepo.GetByTicker(ctx, ticker, s.provider.Source(), period)
}

// isFinancialsFresh returns true if data was fetched within the last 24 hours.
func isFinancialsFresh(latest time.Time) bool {
	if latest.IsZero() {
		return false
	}
	return time.Since(latest) < financialsCacheTTL
}
