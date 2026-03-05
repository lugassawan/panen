package usecase

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

// DividendService provides dividend-specific analytics for DIVIDEND mode portfolios.
type DividendService struct {
	portfolios portfolio.Repository
	holdings   portfolio.HoldingRepository
	stockData  stock.Repository
}

// NewDividendService creates a new DividendService.
func NewDividendService(
	portfolios portfolio.Repository,
	holdings portfolio.HoldingRepository,
	stockData stock.Repository,
) *DividendService {
	return &DividendService{
		portfolios: portfolios,
		holdings:   holdings,
		stockData:  stockData,
	}
}

// GetDividendRanking returns an attractiveness-ranked list of holdings and
// universe candidates for a DIVIDEND mode portfolio.
func (s *DividendService) GetDividendRanking(
	ctx context.Context,
	portfolioID string,
) ([]dividend.RankItem, error) {
	p, err := s.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	if p.Mode != portfolio.ModeDividend {
		return nil, fmt.Errorf("portfolio %s is not in DIVIDEND mode", portfolioID)
	}

	holdings, err := s.holdings.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	thresholds := checklist.ThresholdsForRisk(p.RiskProfile)

	// Compute total portfolio value for position weight calculation.
	type tickerInfo struct {
		data *stock.Data
		val  *valuation.ValuationResult
	}
	infoMap := make(map[string]tickerInfo)
	var totalValue float64

	// Collect data for all held tickers.
	holdingSet := make(map[string]*portfolio.Holding)
	for _, h := range holdings {
		holdingSet[h.Ticker] = h
		data, dataErr := s.stockData.GetByTicker(ctx, h.Ticker)
		if dataErr != nil {
			continue
		}
		val := evaluateStock(data, p.RiskProfile)
		infoMap[h.Ticker] = tickerInfo{data: data, val: val}
		totalValue += data.Price * float64(h.Lots) * 100
	}

	// Collect data for universe tickers not already held.
	for _, ticker := range p.Universe {
		if _, held := holdingSet[ticker]; held {
			continue
		}
		data, dataErr := s.stockData.GetByTicker(ctx, ticker)
		if dataErr != nil {
			continue
		}
		val := evaluateStock(data, p.RiskProfile)
		infoMap[ticker] = tickerInfo{data: data, val: val}
	}

	// Build rank items.
	var items []dividend.RankItem
	for ticker, info := range infoMap {
		h := holdingSet[ticker]
		isHolding := h != nil

		var positionPct float64
		if isHolding && totalValue > 0 {
			positionPct = (info.data.Price * float64(h.Lots) * 100) / totalValue * 100
		}

		var entryPrice, exitTarget float64
		if info.val != nil {
			entryPrice = info.val.EntryPrice
			exitTarget = info.val.ExitTarget
		}

		indicator := dividend.DetermineIndicator(dividend.IndicatorInput{
			HasHolding:     isHolding,
			Price:          info.data.Price,
			EntryPrice:     entryPrice,
			ExitTarget:     exitTarget,
			DividendYield:  info.data.DividendYield,
			PayoutRatio:    info.data.PayoutRatio,
			PositionPct:    positionPct,
			MinDY:          thresholds.MinDY,
			MaxPayoutRatio: thresholds.MaxPayoutRatio,
			MaxPositionPct: thresholds.MaxPositionPct,
		})

		annualDPS := dividend.DeriveAnnualDPS(info.data.Price, info.data.DividendYield)
		var yoc float64
		if isHolding {
			yoc = dividend.YieldOnCost(annualDPS, h.AvgBuyPrice)
		}

		score := dividend.Score(dividend.ScoreInput{
			DY:             info.data.DividendYield,
			MinDY:          thresholds.MinDY,
			PayoutRatio:    info.data.PayoutRatio,
			MaxPayoutRatio: thresholds.MaxPayoutRatio,
			Price:          info.data.Price,
			EntryPrice:     entryPrice,
			PositionPct:    positionPct,
			MaxPositionPct: thresholds.MaxPositionPct,
		})

		items = append(items, dividend.RankItem{
			Ticker:      ticker,
			Indicator:   indicator,
			DY:          info.data.DividendYield,
			YieldOnCost: yoc,
			PayoutRatio: info.data.PayoutRatio,
			PositionPct: positionPct,
			Score:       score,
			IsHolding:   isHolding,
		})
	}

	return dividend.Rank(items), nil
}

func evaluateStock(data *stock.Data, riskProfile portfolio.RiskProfile) *valuation.ValuationResult {
	input := valuation.ValuationInput{
		Ticker:      data.Ticker,
		Price:       data.Price,
		EPS:         data.EPS,
		BVPS:        data.BVPS,
		PBV:         data.PBV,
		PER:         data.PER,
		RiskProfile: valuation.RiskProfile(riskProfile),
	}
	val, err := valuation.Evaluate(input)
	if err != nil {
		return nil
	}
	return val
}
