package presenter

import (
	"context"
	"fmt"

	"github.com/lugassawan/panen/backend/usecase"
)

// PriceHistoryHandler handles price history requests.
type PriceHistoryHandler struct {
	ctx          context.Context
	priceHistory *usecase.PriceHistoryService
}

// NewPriceHistoryHandler creates a new PriceHistoryHandler.
func NewPriceHistoryHandler(ctx context.Context, priceHistory *usecase.PriceHistoryService) *PriceHistoryHandler {
	h := &PriceHistoryHandler{}
	h.Bind(ctx, priceHistory)
	return h
}

func (h *PriceHistoryHandler) Bind(ctx context.Context, priceHistory *usecase.PriceHistoryService) {
	h.ctx = ctx
	h.priceHistory = priceHistory
}

// GetPriceHistory returns historical closing prices for a ticker.
func (h *PriceHistoryHandler) GetPriceHistory(ticker string) ([]PricePointResponse, error) {
	points, err := h.priceHistory.GetHistory(h.ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("get price history: %w", err)
	}
	result := make([]PricePointResponse, len(points))
	for i, p := range points {
		result[i] = PricePointResponse{
			Date:   p.Date.Format(dateLayout),
			Open:   p.Open,
			High:   p.High,
			Low:    p.Low,
			Close:  p.Close,
			Volume: p.Volume,
		}
	}
	return result, nil
}
