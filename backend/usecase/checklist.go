package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

// ChecklistEvaluation is a use-case composite holding auto+manual check results and optional suggestion.
type ChecklistEvaluation struct {
	Action     checklist.ActionType
	Ticker     string
	Checks     []checklist.CheckResult
	AllPassed  bool
	Suggestion *checklist.Suggestion
}

// ChecklistService handles checklist evaluation and manual check persistence.
type ChecklistService struct {
	results    checklist.Repository
	portfolios portfolio.Repository
	holdings   portfolio.HoldingRepository
	brokerages brokerage.Repository
	stockData  stock.Repository
	alertSvc   *AlertService
}

// NewChecklistService creates a new ChecklistService.
func NewChecklistService(
	results checklist.Repository,
	portfolios portfolio.Repository,
	holdings portfolio.HoldingRepository,
	brokerages brokerage.Repository,
	stockData stock.Repository,
	alertSvc *AlertService,
) *ChecklistService {
	return &ChecklistService{
		results:    results,
		portfolios: portfolios,
		holdings:   holdings,
		brokerages: brokerages,
		stockData:  stockData,
		alertSvc:   alertSvc,
	}
}

// Evaluate runs auto-checks and merges saved manual check state, optionally computing a suggestion.
func (s *ChecklistService) Evaluate(
	ctx context.Context,
	portfolioID, ticker string,
	action checklist.ActionType,
) (*ChecklistEvaluation, error) {
	evalInput, err := s.buildEvalInput(ctx, portfolioID, ticker, action)
	if err != nil {
		return nil, err
	}

	autoResults := checklist.EvaluateAutoChecks(evalInput)

	manualResults, err := s.buildManualResults(ctx, portfolioID, ticker, action)
	if err != nil {
		return nil, err
	}

	allChecks := make([]checklist.CheckResult, 0, len(autoResults)+len(manualResults))
	allChecks = append(allChecks, autoResults...)
	allChecks = append(allChecks, manualResults...)

	allPassed := allChecksPassed(allChecks)

	var suggestion *checklist.Suggestion
	if allPassed {
		sug, sugErr := checklist.ComputeSuggestion(evalInput)
		if sugErr != nil && !errors.Is(sugErr, checklist.ErrHoldNoSuggestion) {
			return nil, sugErr
		}
		suggestion = sug
	}

	return &ChecklistEvaluation{
		Action:     action,
		Ticker:     ticker,
		Checks:     allChecks,
		AllPassed:  allPassed,
		Suggestion: suggestion,
	}, nil
}

// ToggleManualCheck persists the completion state of a manual check.
func (s *ChecklistService) ToggleManualCheck(
	ctx context.Context,
	portfolioID, ticker string,
	action checklist.ActionType,
	checkKey string,
	completed bool,
) error {
	result, err := s.results.Get(ctx, portfolioID, ticker, action)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			return err
		}
		result = checklist.NewChecklistResult(portfolioID, ticker, action)
	}

	result.ManualChecks[checkKey] = completed
	result.UpdatedAt = time.Now().UTC()

	return s.results.Upsert(ctx, result)
}

// ResetChecklist deletes any saved checklist result for the given portfolio, ticker, and action.
func (s *ChecklistService) ResetChecklist(
	ctx context.Context,
	portfolioID, ticker string,
	action checklist.ActionType,
) error {
	result, err := s.results.Get(ctx, portfolioID, ticker, action)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.results.Delete(ctx, result.ID)
}

// AvailableActions returns the list of valid actions for a ticker within a portfolio.
func (s *ChecklistService) AvailableActions(
	ctx context.Context,
	portfolioID, ticker string,
) ([]checklist.ActionType, error) {
	p, err := s.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	_, err = s.holdings.GetByPortfolioAndTicker(ctx, portfolioID, ticker)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}
		// No holding found — only BUY is available regardless of mode.
		return []checklist.ActionType{checklist.ActionBuy}, nil
	}

	switch p.Mode {
	case portfolio.ModeValue:
		return []checklist.ActionType{
			checklist.ActionBuy,
			checklist.ActionAverageDown,
			checklist.ActionSellExit,
			checklist.ActionSellStop,
			checklist.ActionHold,
		}, nil
	case portfolio.ModeDividend:
		return []checklist.ActionType{
			checklist.ActionAverageUp,
			checklist.ActionSellExit,
			checklist.ActionSellStop,
			checklist.ActionHold,
		}, nil
	default:
		return []checklist.ActionType{checklist.ActionBuy}, nil
	}
}

// buildEvalInput gathers all data needed for auto-check evaluation.
func (s *ChecklistService) buildEvalInput(
	ctx context.Context,
	portfolioID, ticker string,
	action checklist.ActionType,
) (checklist.EvaluateInput, error) {
	p, err := s.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return checklist.EvaluateInput{}, err
	}

	acct, err := s.brokerages.GetByID(ctx, p.BrokerageAccountID)
	if err != nil {
		return checklist.EvaluateInput{}, err
	}

	data, err := s.fetchStockData(ctx, ticker)
	if err != nil {
		return checklist.EvaluateInput{}, err
	}

	val := computeValuation(data, p)

	holding, err := s.holdings.GetByPortfolioAndTicker(ctx, portfolioID, ticker)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return checklist.EvaluateInput{}, err
	}

	allHoldings, err := s.holdings.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return checklist.EvaluateInput{}, err
	}

	var hasCritical bool
	if s.alertSvc != nil {
		hasCritical, _ = s.alertSvc.HasCriticalAlert(ctx, ticker)
	}

	return checklist.EvaluateInput{
		Action:           action,
		StockData:        data,
		Valuation:        val,
		Holding:          holding,
		Portfolio:        p,
		AllHoldings:      allHoldings,
		Thresholds:       checklist.ThresholdsForRisk(p.RiskProfile),
		BuyFeePct:        acct.BuyFeePct,
		SellFeePct:       acct.SellFeePct,
		SellTaxPct:       acct.SellTaxPct,
		HasCriticalAlert: hasCritical,
	}, nil
}

// fetchStockData retrieves stock data, wrapping not-found as ErrNoStockData.
func (s *ChecklistService) fetchStockData(ctx context.Context, ticker string) (*stock.Data, error) {
	data, err := s.stockData.GetByTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, ErrNoStockData
		}
		return nil, err
	}
	return data, nil
}

// buildManualResults creates manual check results with PENDING status, then merges saved state.
func (s *ChecklistService) buildManualResults(
	ctx context.Context,
	portfolioID, ticker string,
	action checklist.ActionType,
) ([]checklist.CheckResult, error) {
	defs := checklist.ManualCheckDefs(action)
	results := make([]checklist.CheckResult, len(defs))
	for i, def := range defs {
		results[i] = checklist.CheckResult{
			Key:    def.Key,
			Label:  def.Label,
			Type:   checklist.CheckTypeManual,
			Status: checklist.CheckStatusPending,
		}
	}

	saved, err := s.results.Get(ctx, portfolioID, ticker, action)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}
	if saved != nil {
		for i, mr := range results {
			if saved.ManualChecks[mr.Key] {
				results[i].Status = checklist.CheckStatusPass
			}
		}
	}

	return results, nil
}

// computeValuation evaluates a stock's intrinsic value; returns nil on error.
func computeValuation(data *stock.Data, p *portfolio.Portfolio) *valuation.ValuationResult {
	input := valuation.ValuationInput{
		Ticker:      data.Ticker,
		Price:       data.Price,
		EPS:         data.EPS,
		BVPS:        data.BVPS,
		PBV:         data.PBV,
		PER:         data.PER,
		RiskProfile: valuation.RiskProfile(p.RiskProfile),
	}
	val, _ := valuation.Evaluate(input)
	return val
}

// allChecksPassed returns true when every check has status PASS.
func allChecksPassed(checks []checklist.CheckResult) bool {
	for _, cr := range checks {
		if cr.Status != checklist.CheckStatusPass {
			return false
		}
	}
	return true
}
