package usecase

import (
	"context"
	"log"
	"sync"

	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
)

// DividendIncomeSummary holds aggregated dividend income data for a portfolio.
type DividendIncomeSummary struct {
	TotalAnnualIncome float64
	PerStock          []StockIncomeSummary
	MonthlyBreakdown  []dividend.MonthlyIncome
}

// CalendarEntry wraps a projected dividend with the total income for the holding.
type CalendarEntry struct {
	dividend.ProjectedDividend
	TotalIncome float64
}

// StockIncomeSummary holds dividend income data for a single stock in a portfolio.
type StockIncomeSummary struct {
	Ticker       string
	AnnualIncome float64
	DY           float64
	Lots         int
}

// DividendHistoryService handles on-demand fetching and caching of dividend history.
type DividendHistoryService struct {
	historyRepo   dividend.HistoryRepository
	provider      stock.DataProvider
	holdingRepo   portfolio.HoldingRepository
	portfolioRepo portfolio.Repository
	stockRepo     stock.Repository
	fetchMu       sync.Mutex
}

// NewDividendHistoryService creates a new DividendHistoryService.
func NewDividendHistoryService(
	historyRepo dividend.HistoryRepository,
	provider stock.DataProvider,
	holdingRepo portfolio.HoldingRepository,
	portfolioRepo portfolio.Repository,
	stockRepo stock.Repository,
) *DividendHistoryService {
	return &DividendHistoryService{
		historyRepo:   historyRepo,
		provider:      provider,
		holdingRepo:   holdingRepo,
		portfolioRepo: portfolioRepo,
		stockRepo:     stockRepo,
	}
}

// GetDividendHistory returns cached dividend history, refreshing from the provider if stale.
func (s *DividendHistoryService) GetDividendHistory(
	ctx context.Context,
	ticker string,
) ([]dividend.DividendEvent, error) {
	s.fetchMu.Lock()
	defer s.fetchMu.Unlock()

	latest, err := s.historyRepo.LatestDate(ctx, ticker, s.provider.Source())
	if err != nil {
		return nil, err
	}

	if !isFresh(latest) {
		events, err := s.provider.FetchDividendHistory(ctx, ticker)
		if err != nil {
			return nil, err
		}
		if len(events) > 0 {
			if err := s.historyRepo.BulkUpsert(ctx, events); err != nil {
				return nil, err
			}
		}
	}

	return s.historyRepo.GetByTicker(ctx, ticker, s.provider.Source())
}

// GetDGR returns dividend growth rate data for a ticker.
func (s *DividendHistoryService) GetDGR(
	ctx context.Context,
	ticker string,
) ([]dividend.DGRResult, error) {
	events, err := s.GetDividendHistory(ctx, ticker)
	if err != nil {
		return nil, err
	}
	annuals := dividend.AggregateAnnualDPS(events)
	return dividend.CalculateDGR(annuals), nil
}

// GetYoCProgression returns historical yield on cost progression for a holding.
func (s *DividendHistoryService) GetYoCProgression(
	ctx context.Context,
	portfolioID, ticker string,
) ([]dividend.YoCPoint, error) {
	holding, err := s.holdingRepo.GetByPortfolioAndTicker(ctx, portfolioID, ticker)
	if err != nil {
		return nil, err
	}

	events, err := s.GetDividendHistory(ctx, ticker)
	if err != nil {
		return nil, err
	}

	return dividend.YoCProgression(events, holding.AvgBuyPrice), nil
}

// GetDividendIncomeSummary returns aggregated dividend income for a portfolio.
func (s *DividendHistoryService) GetDividendIncomeSummary(
	ctx context.Context,
	portfolioID string,
) (*DividendIncomeSummary, error) {
	holdings, err := s.holdingRepo.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	var totalIncome float64
	var perStock []StockIncomeSummary
	allMonthly := make(map[int]float64)

	for _, h := range holdings {
		events, err := s.GetDividendHistory(ctx, h.Ticker)
		if err != nil {
			log.Printf("warn: failed to fetch dividend history for %s: %v", h.Ticker, err)
			continue
		}

		income := dividend.AnnualDividendIncome(events, h.Lots)
		monthly := dividend.MonthlyDividendIncome(events, h.Lots)

		var dy float64
		stockData, dataErr := s.stockRepo.GetByTicker(ctx, h.Ticker)
		if dataErr == nil {
			dy = stockData.DividendYield
		}

		totalIncome += income
		perStock = append(perStock, StockIncomeSummary{
			Ticker:       h.Ticker,
			AnnualIncome: income,
			DY:           dy,
			Lots:         h.Lots,
		})

		for _, m := range monthly {
			allMonthly[m.Month] += m.Amount
		}
	}

	breakdown := make([]dividend.MonthlyIncome, 0, len(allMonthly))
	for m := 1; m <= 12; m++ {
		if amt, ok := allMonthly[m]; ok {
			breakdown = append(breakdown, dividend.MonthlyIncome{Month: m, Amount: amt})
		}
	}

	return &DividendIncomeSummary{
		TotalAnnualIncome: totalIncome,
		PerStock:          perStock,
		MonthlyBreakdown:  breakdown,
	}, nil
}

// GetDividendCalendar returns projected upcoming dividends for all holdings in a portfolio.
func (s *DividendHistoryService) GetDividendCalendar(
	ctx context.Context,
	portfolioID string,
) ([]CalendarEntry, error) {
	holdings, err := s.holdingRepo.ListByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	sharesPerLot := 100
	var allEntries []CalendarEntry
	for _, h := range holdings {
		events, err := s.GetDividendHistory(ctx, h.Ticker)
		if err != nil {
			log.Printf("warn: failed to fetch dividend history for %s: %v", h.Ticker, err)
			continue
		}

		projections := dividend.ProjectUpcoming(events, h.Ticker)
		for _, p := range projections {
			allEntries = append(allEntries, CalendarEntry{
				ProjectedDividend: p,
				TotalIncome:       p.ExpectedAmount * float64(h.Lots*sharesPerLot),
			})
		}
	}

	return allEntries, nil
}
