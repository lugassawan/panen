package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/usecase"
)

// StockHandler handles stock lookup and valuation requests.
type StockHandler struct {
	ctx    context.Context
	stocks *usecase.StockService
}

// NewStockHandler creates a new StockHandler.
func NewStockHandler(ctx context.Context, stocks *usecase.StockService) *StockHandler {
	return &StockHandler{ctx: ctx, stocks: stocks}
}

// LookupStock fetches or refreshes stock data and returns valuation results.
func (h *StockHandler) LookupStock(ticker, riskProfile string) (*StockValuationResponse, error) {
	rp, err := toValuationRisk(riskProfile)
	if err != nil {
		return nil, err
	}
	data, result, err := h.stocks.Lookup(h.ctx, ticker, rp)
	if err != nil {
		return nil, err
	}
	return buildStockResponse(data, result, riskProfile), nil
}

// GetStockValuation returns cached stock valuation without fetching new data.
func (h *StockHandler) GetStockValuation(ticker, riskProfile string) (*StockValuationResponse, error) {
	rp, err := toValuationRisk(riskProfile)
	if err != nil {
		return nil, err
	}
	data, result, err := h.stocks.GetCached(h.ctx, ticker, rp)
	if err != nil {
		return nil, err
	}
	return buildStockResponse(data, result, riskProfile), nil
}
