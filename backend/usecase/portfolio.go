package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/trailingstop"
	"github.com/lugassawan/panen/backend/domain/valuation"
	"github.com/lugassawan/panen/backend/infra/applog"
)

// HoldingWithValuation is a use-case-layer composite carrying a holding
// together with its optional stock data and valuation result.
type HoldingWithValuation struct {
	Holding         *portfolio.Holding
	StockData       *stock.Data
	Valuation       *valuation.ValuationResult
	TrailingStop    *trailingstop.TrailingStopResult
	DividendMetrics *dividend.DividendMetrics
}

// PortfolioService handles portfolio and holding operations.
type PortfolioService struct {
	portfolios portfolio.Repository
	holdings   portfolio.HoldingRepository
	buyTxns    portfolio.BuyTransactionRepository
	brokerages brokerage.Repository
	stockData  stock.Repository
	peaks      trailingstop.PeakRepository
}

// NewPortfolioService creates a new PortfolioService.
func NewPortfolioService(
	portfolios portfolio.Repository,
	holdings portfolio.HoldingRepository,
	buyTxns portfolio.BuyTransactionRepository,
	brokerages brokerage.Repository,
	stockData stock.Repository,
	peaks trailingstop.PeakRepository,
) *PortfolioService {
	return &PortfolioService{
		portfolios: portfolios,
		holdings:   holdings,
		buyTxns:    buyTxns,
		brokerages: brokerages,
		stockData:  stockData,
		peaks:      peaks,
	}
}

// Create validates and persists a new portfolio.
func (s *PortfolioService) Create(ctx context.Context, p *portfolio.Portfolio) error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmptyName
	}
	// Defense-in-depth: re-validate even though presenter parses these before construction.
	// The usecase is a standalone API boundary callable by future non-presenter callers.
	if _, err := portfolio.ParseMode(string(p.Mode)); err != nil {
		return fmt.Errorf("create portfolio: %w", err)
	}
	if _, err := portfolio.ParseRiskProfile(string(p.RiskProfile)); err != nil {
		return fmt.Errorf("create portfolio: %w", err)
	}

	siblings, err := s.portfolios.ListByBrokerageAccountID(ctx, p.BrokerageAccountID)
	if err != nil {
		return fmt.Errorf("create portfolio: %w", err)
	}
	for _, sib := range siblings {
		if sib.Mode == p.Mode {
			return ErrDuplicateMode
		}
	}

	return s.portfolios.Create(ctx, p)
}

// GetByID returns a single portfolio by ID.
func (s *PortfolioService) GetByID(ctx context.Context, id string) (*portfolio.Portfolio, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyID
	}
	return s.portfolios.GetByID(ctx, id)
}

// Update validates and persists changes to a portfolio.
func (s *PortfolioService) Update(ctx context.Context, p *portfolio.Portfolio) error {
	if strings.TrimSpace(p.ID) == "" {
		return ErrEmptyID
	}
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmptyName
	}
	if _, err := portfolio.ParseRiskProfile(string(p.RiskProfile)); err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}

	existing, err := s.portfolios.GetByID(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}
	if p.Mode != existing.Mode {
		return ErrModeImmutable
	}

	p.UpdatedAt = time.Now().UTC()
	return s.portfolios.Update(ctx, p)
}

// Delete removes a portfolio if it has no holdings.
func (s *PortfolioService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyID
	}
	holdings, err := s.holdings.ListByPortfolioID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete portfolio: %w", err)
	}
	if len(holdings) > 0 {
		return fmt.Errorf("%w: %d holding(s) linked", ErrHasHoldings, len(holdings))
	}
	if err := s.portfolios.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete portfolio: %w", err)
	}
	return nil
}

// ListByBrokerageAccountID returns all portfolios for a brokerage account.
func (s *PortfolioService) ListByBrokerageAccountID(
	ctx context.Context,
	brokerageAccountID string,
) ([]*portfolio.Portfolio, error) {
	return s.portfolios.ListByBrokerageAccountID(ctx, brokerageAccountID)
}

// AddHolding adds or updates a holding within a portfolio, recording a buy transaction.
func (s *PortfolioService) AddHolding(
	ctx context.Context,
	portfolioID, ticker string,
	price float64,
	lots int,
	date time.Time,
) (*portfolio.Holding, error) {
	if strings.TrimSpace(portfolioID) == "" {
		return nil, ErrEmptyID
	}
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if ticker == "" {
		return nil, ErrEmptyTicker
	}
	if price <= 0 {
		return nil, ErrInvalidPrice
	}
	if lots <= 0 {
		return nil, ErrInvalidLots
	}

	p, err := s.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	acct, err := s.brokerages.GetByID(ctx, p.BrokerageAccountID)
	if err != nil {
		return nil, err
	}

	if err := s.checkDuplicateHolding(ctx, p.BrokerageAccountID, portfolioID, ticker); err != nil {
		return nil, err
	}

	fee := portfolio.ComputeBuyFee(price, lots, acct.BuyFeePct)

	existing, err := s.holdings.GetByPortfolioAndTicker(ctx, portfolioID, ticker)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}

	var holding *portfolio.Holding
	if existing != nil {
		existing.AvgBuyPrice = existing.ComputeAvgBuyPrice(price, lots)
		existing.Lots += lots
		existing.UpdatedAt = time.Now().UTC()
		if err := s.holdings.Update(ctx, existing); err != nil {
			return nil, err
		}
		holding = existing
	} else {
		holding = portfolio.NewHolding(portfolioID, ticker, price, lots)
		if err := s.holdings.Create(ctx, holding); err != nil {
			return nil, err
		}
	}

	tx := portfolio.NewBuyTransaction(holding.ID, date, price, lots, fee)
	if err := s.buyTxns.Create(ctx, tx); err != nil {
		return nil, err
	}

	return holding, nil
}

// GetDetail retrieves a portfolio with its holdings and optional valuations.
func (s *PortfolioService) GetDetail(
	ctx context.Context,
	id string,
) (*portfolio.Portfolio, []*HoldingWithValuation, error) {
	p, err := s.portfolios.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	holdings, err := s.holdings.ListByPortfolioID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	isValue := p.Mode == portfolio.ModeValue

	// Batch-fetch existing peaks for VALUE mode portfolios.
	peakMap := make(map[string]*trailingstop.HoldingPeak)
	if isValue && len(holdings) > 0 {
		holdingIDs := make([]string, len(holdings))
		for i, h := range holdings {
			holdingIDs[i] = h.ID
		}
		peaks, peakErr := s.peaks.ListByHoldingIDs(ctx, holdingIDs)
		if peakErr == nil {
			peakMap = shared.IndexBy(peaks, func(pk *trailingstop.HoldingPeak) string { return pk.HoldingID })
		}
	}

	// Pre-fetch stock data for all holdings (avoids N+1 in dividend metrics).
	stockMap := make(map[string]*stock.Data)
	for _, h := range holdings {
		data, dataErr := s.stockData.GetByTicker(ctx, h.Ticker)
		if dataErr == nil {
			stockMap[h.Ticker] = data
		}
	}

	result := make([]*HoldingWithValuation, len(holdings))
	for i, h := range holdings {
		result[i] = enrichHolding(h, p.RiskProfile, isValue, peakMap, holdings, stockMap)
	}

	return p, result, nil
}

func enrichHolding(
	h *portfolio.Holding,
	riskProfile portfolio.RiskProfile,
	isValue bool,
	peakMap map[string]*trailingstop.HoldingPeak,
	allHoldings []*portfolio.Holding,
	stockMap map[string]*stock.Data,
) *HoldingWithValuation {
	hwv := &HoldingWithValuation{Holding: h}

	data := stockMap[h.Ticker]
	if data == nil {
		return hwv
	}

	hwv.StockData = data
	input := valuation.ValuationInput{
		Ticker:      data.Ticker,
		Price:       data.Price,
		EPS:         data.EPS,
		BVPS:        data.BVPS,
		PBV:         data.PBV,
		PER:         data.PER,
		RiskProfile: valuation.RiskProfile(riskProfile),
	}
	val, valErr := valuation.Evaluate(input)
	if valErr == nil {
		hwv.Valuation = val
	}

	if isValue {
		hwv.TrailingStop = computeTrailingStop(h, data, riskProfile, peakMap)
	} else {
		hwv.DividendMetrics = computeDividendMetrics(h, data, val, riskProfile, allHoldings, stockMap)
	}

	return hwv
}

// SyncPeaks updates peak prices for all holdings in a VALUE mode portfolio.
// This is an explicit command that should be called before reading portfolio detail.
func (s *PortfolioService) SyncPeaks(ctx context.Context, portfolioID string) error {
	p, err := s.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("sync peaks: %w", err)
	}
	if p.Mode != portfolio.ModeValue {
		return nil
	}

	holdings, err := s.holdings.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("sync peaks: %w", err)
	}
	if len(holdings) == 0 {
		return nil
	}

	holdingIDs := make([]string, len(holdings))
	for i, h := range holdings {
		holdingIDs[i] = h.ID
	}
	existingPeaks, err := s.peaks.ListByHoldingIDs(ctx, holdingIDs)
	if err != nil {
		return fmt.Errorf("sync peaks: %w", err)
	}
	peakMap := make(map[string]*trailingstop.HoldingPeak, len(existingPeaks))
	for _, pk := range existingPeaks {
		peakMap[pk.HoldingID] = pk
	}

	now := time.Now().UTC()
	for _, h := range holdings {
		data, dataErr := s.stockData.GetByTicker(ctx, h.Ticker)
		if dataErr != nil {
			continue
		}
		s.syncHoldingPeak(ctx, h, data, peakMap, now)
	}
	return nil
}

// syncHoldingPeak is best-effort: peak data enhances trailing-stop display but
// is not critical. Errors are logged and swallowed so a persistence failure
// doesn't block the holdings refresh that the user is waiting on.
func (s *PortfolioService) syncHoldingPeak(
	ctx context.Context,
	h *portfolio.Holding,
	data *stock.Data,
	peakMap map[string]*trailingstop.HoldingPeak,
	now time.Time,
) {
	existing := peakMap[h.ID]
	var currentPeak float64
	if existing != nil {
		currentPeak = existing.PeakPrice
	}
	seedPrice := max(data.Price, data.High52Week)
	newPeak := trailingstop.UpdatePeak(currentPeak, seedPrice)

	if existing == nil {
		peak := &trailingstop.HoldingPeak{
			ID:        shared.NewID(),
			HoldingID: h.ID,
			PeakPrice: newPeak,
			UpdatedAt: now,
		}
		if upsertErr := s.peaks.Upsert(ctx, peak); upsertErr != nil {
			applog.Warn("failed to persist peak", upsertErr, applog.Fields{"holdingID": h.ID})
		}
	} else if newPeak > existing.PeakPrice {
		existing.PeakPrice = newPeak
		existing.UpdatedAt = now
		if upsertErr := s.peaks.Upsert(ctx, existing); upsertErr != nil {
			applog.Warn("failed to update peak", upsertErr, applog.Fields{"holdingID": h.ID})
		}
	}
}

func computeTrailingStop(
	h *portfolio.Holding,
	data *stock.Data,
	riskProfile portfolio.RiskProfile,
	peakMap map[string]*trailingstop.HoldingPeak,
) *trailingstop.TrailingStopResult {
	stopPct, err := trailingstop.StopPercentage(riskProfile)
	if err != nil {
		return nil
	}

	existing := peakMap[h.ID]
	var currentPeak float64
	if existing != nil {
		currentPeak = existing.PeakPrice
	}
	seedPrice := max(data.Price, data.High52Week)
	newPeak := trailingstop.UpdatePeak(currentPeak, seedPrice)

	stopPrice := trailingstop.StopPrice(newPeak, stopPct)
	triggered := trailingstop.IsTriggered(data.Price, stopPrice)
	fundamentals := trailingstop.EvaluateFundamentals(data.ROE, data.DER, data.EPS)

	return &trailingstop.TrailingStopResult{
		PeakPrice:        newPeak,
		StopPct:          stopPct,
		StopPrice:        stopPrice,
		Triggered:        triggered,
		FundamentalExits: fundamentals,
	}
}

func computeDividendMetrics(
	h *portfolio.Holding,
	data *stock.Data,
	val *valuation.ValuationResult,
	riskProfile portfolio.RiskProfile,
	allHoldings []*portfolio.Holding,
	stockMap map[string]*stock.Data,
) *dividend.DividendMetrics {
	thresholds := checklist.ThresholdsForRisk(riskProfile)

	annualDPS := dividend.DeriveAnnualDPS(data.Price, data.DividendYield)
	yoc := dividend.YieldOnCost(annualDPS, h.AvgBuyPrice)
	projectedYoC := dividend.ProjectedYoC(annualDPS, h.AvgBuyPrice, h.Lots, data.Price, 1)

	// Compute portfolio-level yield from pre-fetched stock data.
	var yieldItems []dividend.PortfolioYieldItem
	for _, oh := range allHoldings {
		ohData := stockMap[oh.Ticker]
		if ohData == nil {
			continue
		}
		dps := dividend.DeriveAnnualDPS(ohData.Price, ohData.DividendYield)
		yieldItems = append(yieldItems, dividend.PortfolioYieldItem{
			PositionValue: ohData.Price * float64(oh.Lots) * 100,
			AnnualDPS:     dps,
			Lots:          oh.Lots,
		})
	}
	portfolioYield := dividend.PortfolioYield(yieldItems)

	// Compute position weight for indicator.
	var totalValue float64
	for _, item := range yieldItems {
		totalValue += item.PositionValue
	}
	var positionPct float64
	if totalValue > 0 {
		positionPct = (data.Price * float64(h.Lots) * 100) / totalValue * 100
	}

	var entryPrice, exitTarget float64
	if val != nil {
		entryPrice = val.EntryPrice
		exitTarget = val.ExitTarget
	}

	indicator := dividend.DetermineIndicator(dividend.IndicatorInput{
		HasHolding:     true,
		Price:          data.Price,
		EntryPrice:     entryPrice,
		ExitTarget:     exitTarget,
		DividendYield:  data.DividendYield,
		PayoutRatio:    data.PayoutRatio,
		PositionPct:    positionPct,
		MinDY:          thresholds.MinDY,
		MaxPayoutRatio: thresholds.MaxPayoutRatio,
		MaxPositionPct: thresholds.MaxPositionPct,
	})

	return &dividend.DividendMetrics{
		Indicator:      indicator,
		AnnualDPS:      annualDPS,
		YieldOnCost:    yoc,
		ProjectedYoC:   projectedYoC,
		PortfolioYield: portfolioYield,
	}
}

// checkDuplicateHolding ensures a ticker is not already held in a sibling portfolio
// under the same brokerage account.
func (s *PortfolioService) checkDuplicateHolding(
	ctx context.Context,
	brokerageAccountID, portfolioID, ticker string,
) error {
	siblings, err := s.portfolios.ListByBrokerageAccountID(ctx, brokerageAccountID)
	if err != nil {
		return fmt.Errorf("check duplicate holding: %w", err)
	}
	for _, sib := range siblings {
		if sib.ID == portfolioID {
			continue
		}
		_, sibErr := s.holdings.GetByPortfolioAndTicker(ctx, sib.ID, ticker)
		if sibErr == nil {
			return fmt.Errorf("%w: %s in portfolio %q", ErrDuplicateHolding, ticker, sib.Name)
		}
		if !errors.Is(sibErr, shared.ErrNotFound) {
			return sibErr
		}
	}
	return nil
}
