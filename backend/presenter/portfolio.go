package presenter

import (
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/portfolio"
	"github.com/lugassawan/panen/backend/internal/domain/shared"
)

// CreatePortfolio creates a new portfolio under the given brokerage account.
func (a *App) CreatePortfolio(
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
	if err := a.portfolios.Create(a.ctx, p); err != nil {
		return nil, err
	}
	return buildPortfolioResponse(p), nil
}

// AddHolding adds a stock holding to a portfolio.
func (a *App) AddHolding(
	portfolioID, ticker string,
	price float64,
	lots int,
	dateStr string,
) (*HoldingDetailResponse, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}

	holding, err := a.portfolios.AddHolding(a.ctx, portfolioID, ticker, price, lots, date)
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
func (a *App) GetPortfolio(id string) (*PortfolioDetailResponse, error) {
	p, holdings, err := a.portfolios.GetDetail(a.ctx, id)
	if err != nil {
		return nil, err
	}
	return buildPortfolioDetailResponse(p, holdings), nil
}
