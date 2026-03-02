package presenter

import (
	"context"
	"time"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/shared"
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

// CreatePortfolio creates a new portfolio under the given brokerage account.
func (h *PortfolioHandler) CreatePortfolio(
	brokerageAcctID, name, mode, riskProfile string,
	capital, monthlyAddition float64,
	maxStocks int,
) (*PortfolioResponse, error) {
	m, err := toPortfolioMode(mode)
	if err != nil {
		return nil, err
	}
	rp, err := toPortfolioRisk(riskProfile)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	p := &portfolio.Portfolio{
		ID:                 shared.NewID(),
		BrokerageAccountID: brokerageAcctID,
		Name:               name,
		Mode:               m,
		RiskProfile:        rp,
		Capital:            capital,
		MonthlyAddition:    monthlyAddition,
		MaxStocks:          maxStocks,
		Universe:           []string{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := h.portfolios.Create(h.ctx, p); err != nil {
		return nil, err
	}
	return buildPortfolioResponse(p), nil
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
func (h *PortfolioHandler) GetPortfolio(id string) (*PortfolioDetailResponse, error) {
	p, holdings, err := h.portfolios.GetDetail(h.ctx, id)
	if err != nil {
		return nil, err
	}
	return buildPortfolioDetailResponse(p, holdings), nil
}
