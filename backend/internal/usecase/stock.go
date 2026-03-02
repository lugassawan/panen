package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/shared"
	"github.com/lugassawan/panen/backend/internal/domain/stock"
	"github.com/lugassawan/panen/backend/internal/domain/valuation"
)

const cacheTTL = 24 * time.Hour

// StockService handles stock lookup and valuation.
type StockService struct {
	stockData stock.Repository
	provider  stock.DataProvider
}

// NewStockService creates a new StockService.
func NewStockService(stockData stock.Repository, provider stock.DataProvider) *StockService {
	return &StockService{stockData: stockData, provider: provider}
}

// Lookup fetches stock data (from cache or provider) and computes valuation.
func (s *StockService) Lookup(
	ctx context.Context,
	ticker string,
	rp valuation.RiskProfile,
) (*stock.Data, *valuation.ValuationResult, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if ticker == "" {
		return nil, nil, ErrEmptyTicker
	}

	data, err := s.stockData.GetByTickerAndSource(ctx, ticker, s.provider.Source())
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, nil, err
	}

	if data != nil && time.Since(data.FetchedAt) < cacheTTL {
		result, err := evaluate(data, rp)
		if err != nil {
			return nil, nil, err
		}
		return data, result, nil
	}

	data, err = s.fetch(ctx, ticker)
	if err != nil {
		return nil, nil, err
	}

	if err := s.stockData.Upsert(ctx, data); err != nil {
		return nil, nil, err
	}

	result, err := evaluate(data, rp)
	if err != nil {
		return nil, nil, err
	}
	return data, result, nil
}

// GetCached returns cached stock data and valuation without fetching.
func (s *StockService) GetCached(
	ctx context.Context,
	ticker string,
	rp valuation.RiskProfile,
) (*stock.Data, *valuation.ValuationResult, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if ticker == "" {
		return nil, nil, ErrEmptyTicker
	}

	data, err := s.stockData.GetByTickerAndSource(ctx, ticker, s.provider.Source())
	if errors.Is(err, shared.ErrNotFound) {
		return nil, nil, ErrNoStockData
	}
	if err != nil {
		return nil, nil, err
	}

	result, err := evaluate(data, rp)
	if err != nil {
		return nil, nil, err
	}
	return data, result, nil
}

func (s *StockService) fetch(ctx context.Context, ticker string) (*stock.Data, error) {
	price, err := s.provider.FetchPrice(ctx, ticker)
	if err != nil {
		return nil, err
	}

	fin, err := s.provider.FetchFinancials(ctx, ticker)
	if err != nil {
		return nil, err
	}

	return &stock.Data{
		ID:            shared.NewID(),
		Ticker:        ticker,
		Price:         price.Price,
		High52Week:    price.High52Week,
		Low52Week:     price.Low52Week,
		EPS:           fin.EPS,
		BVPS:          fin.BVPS,
		ROE:           fin.ROE,
		DER:           fin.DER,
		PBV:           fin.PBV,
		PER:           fin.PER,
		DividendYield: fin.DividendYield,
		PayoutRatio:   fin.PayoutRatio,
		FetchedAt:     time.Now().UTC(),
		Source:        s.provider.Source(),
	}, nil
}

func evaluate(data *stock.Data, rp valuation.RiskProfile) (*valuation.ValuationResult, error) {
	input := valuation.ValuationInput{
		Ticker:      data.Ticker,
		Price:       data.Price,
		EPS:         data.EPS,
		BVPS:        data.BVPS,
		PBV:         data.PBV,
		PER:         data.PER,
		RiskProfile: rp,
	}
	return valuation.Evaluate(input)
}
