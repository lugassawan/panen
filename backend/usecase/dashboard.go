package usecase

import (
	"context"
	"errors"
	"sort"

	"github.com/lugassawan/panen/backend/domain/dashboard"
	"github.com/lugassawan/panen/backend/domain/payday"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/transaction"
)

const topMoversLimit = 5

// DashboardService aggregates data across all portfolios for the dashboard overview.
type DashboardService struct {
	portfolios portfolio.Repository
	holdings   portfolio.HoldingRepository
	stocks     stock.Repository
	paydays    payday.Repository
	txnHistory transaction.Repository
	sectorReg  SectorRegistry
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(
	portfolios portfolio.Repository,
	holdings portfolio.HoldingRepository,
	stocks stock.Repository,
	paydays payday.Repository,
	txnHistory transaction.Repository,
	sectorReg SectorRegistry,
) *DashboardService {
	return &DashboardService{
		portfolios: portfolios,
		holdings:   holdings,
		stocks:     stocks,
		paydays:    paydays,
		txnHistory: txnHistory,
		sectorReg:  sectorReg,
	}
}

// GetOverview returns aggregated performance data across all portfolios.
func (s *DashboardService) GetOverview(ctx context.Context) (*dashboard.Overview, error) {
	allPortfolios, err := s.portfolios.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(allPortfolios) == 0 {
		return emptyOverview(), nil
	}

	agg, err := s.aggregatePortfolios(ctx, allPortfolios)
	if err != nil {
		return nil, err
	}

	records, err := s.recentTransactions(ctx)
	if err != nil {
		return nil, err
	}

	gainers, losers := topMovers(agg.allHoldings)

	return &dashboard.Overview{
		TotalMarketValue:    agg.totalMV,
		TotalCostBasis:      agg.totalCB,
		TotalPLAmount:       agg.totalMV - agg.totalCB,
		TotalPLPercent:      plPercent(agg.totalMV, agg.totalCB),
		TotalDividendIncome: agg.totalDividend,
		Portfolios:          agg.summaries,
		TopGainers:          gainers,
		TopLosers:           losers,
		PortfolioAllocation: portfolioAllocation(agg.summaries),
		SectorAllocation:    sectorAllocation(agg.sectorValues, agg.totalMV),
		RecentTransactions:  records,
	}, nil
}

type aggregation struct {
	totalMV       float64
	totalCB       float64
	totalDividend float64
	summaries     []dashboard.PortfolioSummary
	allHoldings   []dashboard.HoldingPL
	sectorValues  map[string]float64
}

func (s *DashboardService) aggregatePortfolios(
	ctx context.Context,
	portfolios []*portfolio.Portfolio,
) (*aggregation, error) {
	agg := &aggregation{sectorValues: make(map[string]float64)}

	for _, p := range portfolios {
		pMV, pCB, err := s.aggregateHoldings(ctx, p, agg)
		if err != nil {
			return nil, err
		}

		agg.summaries = append(agg.summaries, dashboard.PortfolioSummary{
			ID:          p.ID,
			Name:        p.Name,
			Mode:        string(p.Mode),
			MarketValue: pMV,
			CostBasis:   pCB,
			PLAmount:    pMV - pCB,
			PLPercent:   plPercent(pMV, pCB),
		})

		agg.totalMV += pMV
		agg.totalCB += pCB

		div, err := s.confirmedDividendIncome(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		agg.totalDividend += div
	}

	// Compute portfolio weights.
	for i := range agg.summaries {
		if agg.totalMV > 0 {
			agg.summaries[i].Weight = agg.summaries[i].MarketValue / agg.totalMV * 100
		}
	}

	return agg, nil
}

func (s *DashboardService) aggregateHoldings(
	ctx context.Context,
	p *portfolio.Portfolio,
	agg *aggregation,
) (float64, float64, error) {
	holdings, err := s.holdings.ListByPortfolioID(ctx, p.ID)
	if err != nil {
		return 0, 0, err
	}

	var pMV, pCB float64
	for _, h := range holdings {
		price, err := s.currentPrice(ctx, h)
		if err != nil {
			return 0, 0, err
		}

		mv := price * float64(h.Lots) * 100
		cb := h.AvgBuyPrice * float64(h.Lots) * 100
		pMV += mv
		pCB += cb

		agg.allHoldings = append(agg.allHoldings, dashboard.HoldingPL{
			Ticker:        h.Ticker,
			PortfolioID:   p.ID,
			PortfolioName: p.Name,
			MarketValue:   mv,
			CostBasis:     cb,
			PLAmount:      mv - cb,
			PLPercent:     plPercent(mv, cb),
		})

		sector := s.sectorReg.SectorOf(h.Ticker)
		if sector == "" {
			sector = "Other"
		}
		agg.sectorValues[sector] += mv
	}

	return pMV, pCB, nil
}

func (s *DashboardService) currentPrice(ctx context.Context, h *portfolio.Holding) (float64, error) {
	sd, err := s.stocks.GetByTicker(ctx, h.Ticker)
	if err == nil {
		return sd.Price, nil
	}
	if errors.Is(err, shared.ErrNotFound) {
		return h.AvgBuyPrice, nil
	}
	return 0, err
}

func (s *DashboardService) confirmedDividendIncome(ctx context.Context, portfolioID string) (float64, error) {
	events, err := s.paydays.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, ev := range events {
		if ev.Status == payday.StatusConfirmed {
			total += ev.Actual
		}
	}
	return total, nil
}

func (s *DashboardService) recentTransactions(ctx context.Context) ([]transaction.Record, error) {
	records, err := s.txnHistory.List(ctx, transaction.Filter{SortField: "date"})
	if err != nil {
		return nil, err
	}
	if len(records) > 10 {
		records = records[:10]
	}
	return records, nil
}

func emptyOverview() *dashboard.Overview {
	return &dashboard.Overview{
		Portfolios:          []dashboard.PortfolioSummary{},
		TopGainers:          []dashboard.HoldingPL{},
		TopLosers:           []dashboard.HoldingPL{},
		PortfolioAllocation: []dashboard.AllocationItem{},
		SectorAllocation:    []dashboard.AllocationItem{},
		RecentTransactions:  []transaction.Record{},
	}
}

func plPercent(marketValue, costBasis float64) float64 {
	if costBasis <= 0 {
		return 0
	}
	return (marketValue - costBasis) / costBasis * 100
}

func portfolioAllocation(summaries []dashboard.PortfolioSummary) []dashboard.AllocationItem {
	items := make([]dashboard.AllocationItem, len(summaries))
	for i, ps := range summaries {
		items[i] = dashboard.AllocationItem{
			Label: ps.Name,
			Value: ps.MarketValue,
			Pct:   ps.Weight,
		}
	}
	return items
}

func sectorAllocation(values map[string]float64, totalMV float64) []dashboard.AllocationItem {
	items := make([]dashboard.AllocationItem, 0, len(values))
	for sector, value := range values {
		var pct float64
		if totalMV > 0 {
			pct = value / totalMV * 100
		}
		items = append(items, dashboard.AllocationItem{
			Label: sector,
			Value: value,
			Pct:   pct,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})
	return items
}

func topMovers(all []dashboard.HoldingPL) (gainers, losers []dashboard.HoldingPL) {
	var pos, neg []dashboard.HoldingPL
	for _, h := range all {
		if h.PLAmount > 0 {
			pos = append(pos, h)
		} else if h.PLAmount < 0 {
			neg = append(neg, h)
		}
	}

	sort.Slice(pos, func(i, j int) bool { return pos[i].PLPercent > pos[j].PLPercent })
	sort.Slice(neg, func(i, j int) bool { return neg[i].PLPercent < neg[j].PLPercent })

	if len(pos) > topMoversLimit {
		pos = pos[:topMoversLimit]
	}
	if len(neg) > topMoversLimit {
		neg = neg[:topMoversLimit]
	}

	return pos, neg
}
