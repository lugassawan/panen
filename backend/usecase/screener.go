package usecase

import (
	"context"
	"sort"

	"github.com/lugassawan/panen/backend/domain/screener"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

// UniverseType indicates the source of tickers for screening.
type UniverseType string

const (
	UniverseIndex  UniverseType = "INDEX"
	UniverseSector UniverseType = "SECTOR"
	UniverseCustom UniverseType = "CUSTOM"
)

// ScreenRequest holds parameters for running a stock screen.
type ScreenRequest struct {
	UniverseType  UniverseType
	UniverseName  string
	CustomTickers []string
	RiskProfile   string
	SectorFilter  string
	SortField     string
	SortAsc       bool
}

// ScreenResult holds the screening outcome for a single stock.
type ScreenResult struct {
	Ticker    string
	Sector    string
	StockData *stock.Data
	Valuation *valuation.ValuationResult
	Checks    []screener.Check
	Passed    bool
	Score     float64
}

// ScreenerService handles stock screening operations.
type ScreenerService struct {
	stockData      stock.Repository
	indexRegistry  IndexRegistry
	sectorRegistry SectorRegistry
}

// NewScreenerService creates a new ScreenerService.
func NewScreenerService(
	stockData stock.Repository,
	indexRegistry IndexRegistry,
	sectorRegistry SectorRegistry,
) *ScreenerService {
	return &ScreenerService{
		stockData:      stockData,
		indexRegistry:  indexRegistry,
		sectorRegistry: sectorRegistry,
	}
}

// Screen runs the screener against the requested universe.
func (s *ScreenerService) Screen(ctx context.Context, req ScreenRequest) ([]*ScreenResult, error) {
	rp, err := valuation.ParseRiskProfile(req.RiskProfile)
	if err != nil {
		return nil, err
	}

	tickers, err := s.resolveTickers(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(tickers) == 0 {
		return nil, ErrEmptyUniverse
	}

	criteria := screener.CriteriaForRisk(rp)
	results := make([]*ScreenResult, 0, len(tickers))

	for _, ticker := range tickers {
		sector := s.sectorRegistry.SectorOf(ticker)
		if req.SectorFilter != "" && sector != req.SectorFilter {
			continue
		}

		result := &ScreenResult{
			Ticker: ticker,
			Sector: sector,
		}

		data, dataErr := s.stockData.GetByTicker(ctx, ticker)
		if dataErr == nil {
			result.StockData = data

			input := valuation.ValuationInput{
				Ticker:      data.Ticker,
				Price:       data.Price,
				EPS:         data.EPS,
				BVPS:        data.BVPS,
				PBV:         data.PBV,
				PER:         data.PER,
				RiskProfile: rp,
			}
			val, valErr := valuation.Evaluate(input)
			if valErr == nil {
				result.Valuation = val
			}

			eval := screener.Evaluate(data, criteria, val)
			if eval != nil {
				result.Checks = eval.Checks
				result.Passed = eval.Passed
				result.Score = eval.Score
			}
		}

		results = append(results, result)
	}

	sortResults(results, req.SortField, req.SortAsc)

	return results, nil
}

// ListIndexNames returns all registered index names.
func (s *ScreenerService) ListIndexNames() []string {
	return s.indexRegistry.Names()
}

// ListSectors returns all unique sector names.
func (s *ScreenerService) ListSectors() []string {
	return s.sectorRegistry.AllSectors()
}

func (s *ScreenerService) resolveTickers(ctx context.Context, req ScreenRequest) ([]string, error) {
	switch req.UniverseType {
	case UniverseIndex:
		tickers, ok := s.indexRegistry.Tickers(req.UniverseName)
		if !ok {
			return nil, ErrUnknownIndex
		}
		return tickers, nil
	case UniverseSector:
		allTickers, err := s.stockData.ListAllTickers(ctx)
		if err != nil {
			return nil, err
		}
		var filtered []string
		for _, t := range allTickers {
			if s.sectorRegistry.SectorOf(t) == req.UniverseName {
				filtered = append(filtered, t)
			}
		}
		return filtered, nil
	case UniverseCustom:
		return req.CustomTickers, nil
	default:
		return nil, ErrEmptyUniverse
	}
}

func sortResults(results []*ScreenResult, field string, asc bool) {
	sort.SliceStable(results, func(i, j int) bool {
		vi := sortValue(results[i], field)
		vj := sortValue(results[j], field)
		if asc {
			return vi < vj
		}
		return vi > vj
	})
}

func sortValue(r *ScreenResult, field string) float64 {
	if r.StockData == nil {
		return -1e18
	}
	switch field {
	case "price":
		return r.StockData.Price
	case "roe":
		return r.StockData.ROE
	case "der":
		return r.StockData.DER
	case "dividendYield":
		return r.StockData.DividendYield
	default:
		return r.Score
	}
}
