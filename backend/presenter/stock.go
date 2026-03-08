package presenter

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/domain/valuation"
	"github.com/lugassawan/panen/backend/usecase"
)

// StockHandler handles stock lookup and valuation requests.
type StockHandler struct {
	ctx    context.Context
	stocks *usecase.StockService
}

// NewStockHandler creates a new StockHandler.
func NewStockHandler(ctx context.Context, stocks *usecase.StockService) *StockHandler {
	h := &StockHandler{}
	h.Bind(ctx, stocks)
	return h
}

func (h *StockHandler) Bind(ctx context.Context, stocks *usecase.StockService) {
	h.ctx = ctx
	h.stocks = stocks
}

// LookupStock fetches or refreshes stock data and returns valuation results.
func (h *StockHandler) LookupStock(ticker, riskProfile string) (*StockValuationResponse, error) {
	rp, err := valuation.ParseRiskProfile(riskProfile)
	if err != nil {
		return nil, fmt.Errorf("lookup stock: %w", err)
	}
	data, result, err := h.stocks.Lookup(h.ctx, ticker, rp)
	if err != nil {
		return nil, fmt.Errorf("lookup stock: %w", err)
	}
	return newStockValuationResponse(data, result, riskProfile), nil
}

// GetStockValuation returns cached stock valuation without fetching new data.
func (h *StockHandler) GetStockValuation(ticker, riskProfile string) (*StockValuationResponse, error) {
	rp, err := valuation.ParseRiskProfile(riskProfile)
	if err != nil {
		return nil, fmt.Errorf("get stock valuation: %w", err)
	}
	data, result, err := h.stocks.GetCached(h.ctx, ticker, rp)
	if err != nil {
		return nil, fmt.Errorf("get stock valuation: %w", err)
	}
	return newStockValuationResponse(data, result, riskProfile), nil
}
