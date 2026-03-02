package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

// HoldingWithValuation is a use-case-layer composite carrying a holding
// together with its optional stock data and valuation result.
type HoldingWithValuation struct {
	Holding   *portfolio.Holding
	StockData *stock.Data
	Valuation *valuation.ValuationResult
}

// PortfolioService handles portfolio and holding operations.
type PortfolioService struct {
	portfolios portfolio.Repository
	holdings   portfolio.HoldingRepository
	buyTxns    portfolio.BuyTransactionRepository
	brokerages brokerage.Repository
	stockData  stock.Repository
}

// NewPortfolioService creates a new PortfolioService.
func NewPortfolioService(
	portfolios portfolio.Repository,
	holdings portfolio.HoldingRepository,
	buyTxns portfolio.BuyTransactionRepository,
	brokerages brokerage.Repository,
	stockData stock.Repository,
) *PortfolioService {
	return &PortfolioService{
		portfolios: portfolios,
		holdings:   holdings,
		buyTxns:    buyTxns,
		brokerages: brokerages,
		stockData:  stockData,
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
	return s.portfolios.Create(ctx, p)
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

	shares := float64(lots) * 100 // 1 lot = 100 shares on IDX
	fee := price * shares * acct.BuyFeePct / 100

	existing, err := s.holdings.GetByPortfolioAndTicker(ctx, portfolioID, ticker)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}

	var holding *portfolio.Holding
	if existing != nil {
		totalLots := existing.Lots + lots
		totalCost := existing.AvgBuyPrice*float64(existing.Lots) + price*float64(lots)
		existing.AvgBuyPrice = totalCost / float64(totalLots)
		existing.Lots = totalLots
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

	result := make([]*HoldingWithValuation, len(holdings))
	for i, h := range holdings {
		hwv := &HoldingWithValuation{Holding: h}

		data, err := s.stockData.GetByTicker(ctx, h.Ticker)
		if err == nil {
			hwv.StockData = data
			input := valuation.ValuationInput{
				Ticker:      data.Ticker,
				Price:       data.Price,
				EPS:         data.EPS,
				BVPS:        data.BVPS,
				PBV:         data.PBV,
				PER:         data.PER,
				RiskProfile: valuation.RiskProfile(p.RiskProfile),
			}
			val, valErr := valuation.Evaluate(input)
			if valErr == nil {
				hwv.Valuation = val
			}
		}

		result[i] = hwv
	}

	return p, result, nil
}
