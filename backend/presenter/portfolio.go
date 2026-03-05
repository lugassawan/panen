package presenter

import (
	"context"
	"time"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/usecase"
)

// PortfolioHandler handles portfolio management requests.
type PortfolioHandler struct {
	ctx        context.Context
	portfolios *usecase.PortfolioService
}

// NewPortfolioHandler creates a new PortfolioHandler.
func NewPortfolioHandler(ctx context.Context, portfolios *usecase.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{ctx: ctx, portfolios: portfolios}
}

// ListPortfolios returns all portfolios for a brokerage account.
func (h *PortfolioHandler) ListPortfolios(brokerageAcctID string) ([]*PortfolioResponse, error) {
	portfolios, err := h.portfolios.ListByBrokerageAccountID(h.ctx, brokerageAcctID)
	if err != nil {
		return nil, err
	}
	result := make([]*PortfolioResponse, len(portfolios))
	for i, p := range portfolios {
		result[i] = newPortfolioResponse(p)
	}
	return result, nil
}

// CreatePortfolio creates a new portfolio under the given brokerage account.
func (h *PortfolioHandler) CreatePortfolio(
	brokerageAcctID, name, mode, riskProfile string,
	capital, monthlyAddition float64,
	maxStocks int,
) (*PortfolioResponse, error) {
	m, err := portfolio.ParseMode(mode)
	if err != nil {
		return nil, err
	}
	rp, err := portfolio.ParseRiskProfile(riskProfile)
	if err != nil {
		return nil, err
	}

	p := portfolio.NewPortfolio(brokerageAcctID, name, m, rp, capital, monthlyAddition, maxStocks)
	if err := h.portfolios.Create(h.ctx, p); err != nil {
		return nil, err
	}
	return newPortfolioResponse(p), nil
}

// AddHolding adds a stock holding to a portfolio.
func (h *PortfolioHandler) AddHolding(
	portfolioID, ticker string,
	price float64,
	lots int,
	dateStr string,
) (*HoldingDetailResponse, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}

	holding, err := h.portfolios.AddHolding(h.ctx, portfolioID, ticker, price, lots, date)
	if err != nil {
		return nil, err
	}

	resp := HoldingDetailResponse{
		ID:          holding.ID,
		Ticker:      holding.Ticker,
		AvgBuyPrice: holding.AvgBuyPrice,
		Lots:        holding.Lots,
	}
	return &resp, nil
}

// GetPortfolio returns a portfolio with all holdings and optional valuations.
// SyncPeaks is called first to update trailing stop peak prices (command),
// followed by GetDetail to read the portfolio (query).
func (h *PortfolioHandler) GetPortfolio(id string) (*PortfolioDetailResponse, error) {
	if err := h.portfolios.SyncPeaks(h.ctx, id); err != nil {
		return nil, err
	}
	p, holdings, err := h.portfolios.GetDetail(h.ctx, id)
	if err != nil {
		return nil, err
	}
	return newPortfolioDetailResponse(p, holdings), nil
}

// UpdatePortfolio updates a portfolio's mutable fields (mode is locked post-creation).
func (h *PortfolioHandler) UpdatePortfolio(
	id, name, riskProfile string,
	capital, monthlyAddition float64,
	maxStocks int,
) (*PortfolioResponse, error) {
	rp, err := portfolio.ParseRiskProfile(riskProfile)
	if err != nil {
		return nil, err
	}
	p, err := h.portfolios.GetByID(h.ctx, id)
	if err != nil {
		return nil, err
	}
	p.Name = name
	p.RiskProfile = rp
	p.Capital = capital
	p.MonthlyAddition = monthlyAddition
	p.MaxStocks = maxStocks
	if err := h.portfolios.Update(h.ctx, p); err != nil {
		return nil, err
	}
	return newPortfolioResponse(p), nil
}

// DeletePortfolio removes a portfolio by ID.
func (h *PortfolioHandler) DeletePortfolio(id string) error {
	return h.portfolios.Delete(h.ctx, id)
}
