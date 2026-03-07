package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/usecase"
)

// DividendCalendarHandler handles dividend calendar and history requests.
type DividendCalendarHandler struct {
	ctx     context.Context
	divHist *usecase.DividendHistoryService
}

// NewDividendCalendarHandler creates a new DividendCalendarHandler.
func NewDividendCalendarHandler(
	ctx context.Context,
	divHist *usecase.DividendHistoryService,
) *DividendCalendarHandler {
	h := &DividendCalendarHandler{}
	h.Bind(ctx, divHist)
	return h
}

func (h *DividendCalendarHandler) Bind(ctx context.Context, divHist *usecase.DividendHistoryService) {
	h.ctx = ctx
	h.divHist = divHist
}

// GetDividendHistory returns historical dividend events for a ticker.
func (h *DividendCalendarHandler) GetDividendHistory(ticker string) ([]DividendHistoryItemResponse, error) {
	events, err := h.divHist.GetDividendHistory(h.ctx, ticker)
	if err != nil {
		return nil, err
	}
	result := make([]DividendHistoryItemResponse, len(events))
	for i, e := range events {
		result[i] = DividendHistoryItemResponse{
			ExDate: e.ExDate.Format(dateLayout),
			Amount: e.Amount,
		}
	}
	return result, nil
}

// GetDividendIncomeSummary returns aggregated dividend income for a portfolio.
func (h *DividendCalendarHandler) GetDividendIncomeSummary(portfolioID string) (*DividendIncomeSummaryResponse, error) {
	summary, err := h.divHist.GetDividendIncomeSummary(h.ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	perStock := make([]StockIncomeItemResponse, len(summary.PerStock))
	for i, s := range summary.PerStock {
		perStock[i] = StockIncomeItemResponse{
			Ticker:        s.Ticker,
			AnnualIncome:  s.AnnualIncome,
			DividendYield: s.DY,
			Lots:          s.Lots,
		}
	}

	monthly := make([]MonthlyIncomeItemResponse, len(summary.MonthlyBreakdown))
	for i, m := range summary.MonthlyBreakdown {
		monthly[i] = MonthlyIncomeItemResponse{
			Month:  m.Month,
			Amount: m.Amount,
		}
	}

	return &DividendIncomeSummaryResponse{
		TotalAnnualIncome: summary.TotalAnnualIncome,
		PerStock:          perStock,
		MonthlyBreakdown:  monthly,
	}, nil
}

// GetDGR returns dividend growth rate data for a ticker.
func (h *DividendCalendarHandler) GetDGR(ticker string) ([]DGRItemResponse, error) {
	results, err := h.divHist.GetDGR(h.ctx, ticker)
	if err != nil {
		return nil, err
	}
	resp := make([]DGRItemResponse, len(results))
	for i, r := range results {
		resp[i] = DGRItemResponse{
			Year:      r.Year,
			DPS:       r.DPS,
			GrowthPct: r.GrowthPct,
		}
	}
	return resp, nil
}

// GetYoCProgression returns historical YoC progression for a holding.
func (h *DividendCalendarHandler) GetYoCProgression(portfolioID, ticker string) ([]YoCPointResponse, error) {
	points, err := h.divHist.GetYoCProgression(h.ctx, portfolioID, ticker)
	if err != nil {
		return nil, err
	}
	resp := make([]YoCPointResponse, len(points))
	for i, p := range points {
		resp[i] = YoCPointResponse{
			Date: p.Date.Format(dateLayout),
			YoC:  p.YoC,
		}
	}
	return resp, nil
}

// GetDividendCalendar returns projected upcoming dividends for a portfolio.
func (h *DividendCalendarHandler) GetDividendCalendar(portfolioID string) ([]DividendCalendarEntryResponse, error) {
	projections, err := h.divHist.GetDividendCalendar(h.ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	resp := make([]DividendCalendarEntryResponse, len(projections))
	for i, p := range projections {
		resp[i] = DividendCalendarEntryResponse{
			Ticker:       p.Ticker,
			ExDate:       p.ExpectedExDate.Format(dateLayout),
			Amount:       p.ExpectedAmount,
			IsProjection: p.IsProjection,
			TotalIncome:  p.TotalIncome,
		}
	}
	return resp, nil
}
