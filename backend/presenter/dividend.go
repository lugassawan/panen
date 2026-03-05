package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/usecase"
)

// DividendHandler handles dividend analytics requests.
type DividendHandler struct {
	ctx       context.Context
	dividends *usecase.DividendService
}

// NewDividendHandler creates a new DividendHandler.
func NewDividendHandler(ctx context.Context, dividends *usecase.DividendService) *DividendHandler {
	return &DividendHandler{ctx: ctx, dividends: dividends}
}

// GetDividendRanking returns ranked dividend stocks for a portfolio.
func (h *DividendHandler) GetDividendRanking(portfolioID string) ([]DividendRankItemResponse, error) {
	items, err := h.dividends.GetDividendRanking(h.ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	result := make([]DividendRankItemResponse, len(items))
	for i, item := range items {
		result[i] = newDividendRankItemResponse(item)
	}
	return result, nil
}
