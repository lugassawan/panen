package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		return err
	}
	if _, err := portfolio.ParseRiskProfile(string(p.RiskProfile)); err != nil {
		return err
	}

	siblings, err := s.portfolios.ListByBrokerageAccountID(ctx, p.BrokerageAccountID)
	if err != nil {
		return err
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
		return err
	}

	existing, err := s.portfolios.GetByID(ctx, p.ID)
	if err != nil {
		return err
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
		return err
	}
	if len(holdings) > 0 {
		return fmt.Errorf("%w: %d holding(s) linked", ErrHasHoldings, len(holdings))
	}
	return s.portfolios.Delete(ctx, id)
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

	// Batch-fetch existing peaks for VALUE mode portfolios.
	peakMap := make(map[string]*trailingstop.HoldingPeak)
	isValue := p.Mode == portfolio.ModeValue
	if isValue && len(holdings) > 0 {
		holdingIDs := make([]string, len(holdings))
		for i, h := range holdings {
			holdingIDs[i] = h.ID
		}
		peaks, peakErr := s.peaks.ListByHoldingIDs(ctx, holdingIDs)
		if peakErr == nil {
			for _, pk := range peaks {
				peakMap[pk.HoldingID] = pk
			}
		}
	}

	result := make([]*HoldingWithValuation, len(holdings))
	for i, h := range holdings {
		result[i] = s.enrichHolding(ctx, h, p.RiskProfile, isValue, peakMap, holdings)
	}

	return p, result, nil
}

func (s *PortfolioService) enrichHolding(
	ctx context.Context,
	h *portfolio.Holding,
	riskProfile portfolio.RiskProfile,
	isValue bool,
	peakMap map[string]*trailingstop.HoldingPeak,
	allHoldings []*portfolio.Holding,
) *HoldingWithValuation {
	hwv := &HoldingWithValuation{Holding: h}

	data, err := s.stockData.GetByTicker(ctx, h.Ticker)
	if err != nil {
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
		hwv.TrailingStop = s.computeTrailingStop(ctx, h, data, riskProfile, peakMap)
	} else {
		hwv.DividendMetrics = s.computeDividendMetrics(ctx, h, data, val, riskProfile, allHoldings)
	}

	return hwv
}

func (s *PortfolioService) computeTrailingStop(
	ctx context.Context,
	h *portfolio.Holding,
	data *stock.Data,
	riskProfile portfolio.RiskProfile,
	peakMap map[string]*trailingstop.HoldingPeak,
) *trailingstop.TrailingStopResult {
	stopPct, err := trailingstop.StopPercentage(riskProfile)
	if err != nil {
		return nil
	}

	// Determine current peak. Seed with High52Week to avoid too-tight stops
	// when the feature is first activated on an existing holding.
	existing := peakMap[h.ID]
	var currentPeak float64
	if existing != nil {
		currentPeak = existing.PeakPrice
	}
	seedPrice := max(data.Price, data.High52Week)
	newPeak := trailingstop.UpdatePeak(currentPeak, seedPrice)

	// Persist updated peak (side-effect in read path for lazy peak tracking).
	now := time.Now().UTC()
	if existing == nil {
		peak := &trailingstop.HoldingPeak{
			ID:        shared.NewID(),
			HoldingID: h.ID,
			PeakPrice: newPeak,
			UpdatedAt: now,
		}
		if upsertErr := s.peaks.Upsert(ctx, peak); upsertErr != nil {
			log.Printf("warn: failed to persist peak for holding %s: %v", h.ID, upsertErr)
		}
	} else if newPeak > existing.PeakPrice {
		existing.PeakPrice = newPeak
		existing.UpdatedAt = now
		if upsertErr := s.peaks.Upsert(ctx, existing); upsertErr != nil {
			log.Printf("warn: failed to update peak for holding %s: %v", h.ID, upsertErr)
		}
	}

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

func (s *PortfolioService) computeDividendMetrics(
	ctx context.Context,
	h *portfolio.Holding,
	data *stock.Data,
	val *valuation.ValuationResult,
	riskProfile portfolio.RiskProfile,
	allHoldings []*portfolio.Holding,
) *dividend.DividendMetrics {
	thresholds := checklist.ThresholdsForRisk(riskProfile)

	annualDPS := dividend.DeriveAnnualDPS(data.Price, data.DividendYield)
	yoc := dividend.YieldOnCost(annualDPS, h.AvgBuyPrice)
	projectedYoC := dividend.ProjectedYoC(annualDPS, h.AvgBuyPrice, h.Lots, data.Price, 1)

	// Compute portfolio-level yield from all holdings with available stock data.
	var yieldItems []dividend.PortfolioYieldItem
	for _, oh := range allHoldings {
		ohData, err := s.stockData.GetByTicker(ctx, oh.Ticker)
		if err != nil {
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
		return err
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
