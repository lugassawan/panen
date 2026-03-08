package presenter

import (
	"context"
	"fmt"
	"time"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/infra/applog"
	"github.com/lugassawan/panen/backend/usecase"
)

// PortfolioHandler handles portfolio management requests.
type PortfolioHandler struct {
	ctx             context.Context
	portfolios      *usecase.PortfolioService
	sectors         usecase.SectorRegistry
	preDeleteBackup func(label string) error
}

// NewPortfolioHandler creates a new PortfolioHandler.
func NewPortfolioHandler(
	ctx context.Context,
	portfolios *usecase.PortfolioService,
	sectors usecase.SectorRegistry,
) *PortfolioHandler {
	h := &PortfolioHandler{}
	h.Bind(ctx, portfolios, sectors)
	return h
}

func (h *PortfolioHandler) Bind(
	ctx context.Context,
	portfolios *usecase.PortfolioService,
	sectors usecase.SectorRegistry,
) {
	h.ctx = ctx
	h.portfolios = portfolios
	h.sectors = sectors
}

// BindBackup injects the pre-destructive backup callback.
func (h *PortfolioHandler) BindBackup(fn func(label string) error) {
	h.preDeleteBackup = fn
}

// GetHoldingSectors returns the sector for each ticker.
func (h *PortfolioHandler) GetHoldingSectors(tickers []string) map[string]string {
	result := make(map[string]string, len(tickers))
	for _, t := range tickers {
		result[t] = h.sectors.SectorOf(t)
	}
	return result
}

// ListPortfolios returns all portfolios for a brokerage account.
func (h *PortfolioHandler) ListPortfolios(brokerageAcctID string) ([]*PortfolioResponse, error) {
	portfolios, err := h.portfolios.ListByBrokerageAccountID(h.ctx, brokerageAcctID)
	if err != nil {
		return nil, fmt.Errorf("list portfolios: %w", err)
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
		return nil, fmt.Errorf("create portfolio: %w", err)
	}
	rp, err := portfolio.ParseRiskProfile(riskProfile)
	if err != nil {
		return nil, fmt.Errorf("create portfolio: %w", err)
	}

	p := portfolio.NewPortfolio(brokerageAcctID, name, m, rp, capital, monthlyAddition, maxStocks)
	if err := h.portfolios.Create(h.ctx, p); err != nil {
		return nil, toAppError(fmt.Errorf("create portfolio: %w", err))
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
	date, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return nil, fmt.Errorf("add holding: %w", err)
	}

	holding, err := h.portfolios.AddHolding(h.ctx, portfolioID, ticker, price, lots, date)
	if err != nil {
		return nil, toAppError(fmt.Errorf("add holding: %w", err))
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
		return nil, fmt.Errorf("get portfolio: %w", err)
	}
	p, holdings, err := h.portfolios.GetDetail(h.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get portfolio: %w", err)
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
		return nil, fmt.Errorf("update portfolio: %w", err)
	}
	p, err := h.portfolios.GetByID(h.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("update portfolio: %w", err)
	}
	p.Name = name
	p.RiskProfile = rp
	p.Capital = capital
	p.MonthlyAddition = monthlyAddition
	p.MaxStocks = maxStocks
	if err := h.portfolios.Update(h.ctx, p); err != nil {
		return nil, toAppError(fmt.Errorf("update portfolio: %w", err))
	}
	return newPortfolioResponse(p), nil
}

// DeletePortfolio removes a portfolio by ID.
// A pre-destructive backup is attempted before deletion (non-fatal on failure).
func (h *PortfolioHandler) DeletePortfolio(id string) error {
	if h.preDeleteBackup != nil {
		if err := h.preDeleteBackup("delete"); err != nil {
			applog.Warn("pre-delete backup failed", err, applog.Fields{"portfolioID": id})
		}
	}
	if err := h.portfolios.Delete(h.ctx, id); err != nil {
		return toAppError(fmt.Errorf("delete portfolio: %w", err))
	}
	return nil
}
