package presenter

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
	"github.com/lugassawan/panen/backend/usecase"
)

const timeLayout = "2006-01-02T15:04:05Z"

// --- Domain → DTO constructors (presentation concern) ---

func newStockValuationResponse(
	data *stock.Data,
	result *valuation.ValuationResult,
	riskProfile string,
) *StockValuationResponse {
	return &StockValuationResponse{
		Ticker:         data.Ticker,
		Price:          data.Price,
		High52Week:     data.High52Week,
		Low52Week:      data.Low52Week,
		EPS:            data.EPS,
		BVPS:           data.BVPS,
		ROE:            data.ROE,
		DER:            data.DER,
		PBV:            data.PBV,
		PER:            data.PER,
		DividendYield:  data.DividendYield,
		PayoutRatio:    data.PayoutRatio,
		GrahamNumber:   result.GrahamNumber,
		MarginOfSafety: result.MarginOfSafety,
		EntryPrice:     result.EntryPrice,
		ExitTarget:     result.ExitTarget,
		Verdict:        string(result.Verdict),
		RiskProfile:    riskProfile,
		FetchedAt:      formatDTO(data.FetchedAt),
		Source:         data.Source,
	}
}

func newBrokerageAccountResponse(acct *brokerage.Account) *BrokerageAccountResponse {
	return &BrokerageAccountResponse{
		ID:          acct.ID,
		BrokerName:  acct.BrokerName,
		BuyFeePct:   acct.BuyFeePct,
		SellFeePct:  acct.SellFeePct,
		IsManualFee: acct.IsManualFee,
		CreatedAt:   formatDTO(acct.CreatedAt),
		UpdatedAt:   formatDTO(acct.UpdatedAt),
	}
}

func newPortfolioResponse(p *portfolio.Portfolio) *PortfolioResponse {
	return &PortfolioResponse{
		ID:              p.ID,
		BrokerageAcctID: p.BrokerageAccountID,
		Name:            p.Name,
		Mode:            string(p.Mode),
		RiskProfile:     string(p.RiskProfile),
		Capital:         p.Capital,
		MonthlyAddition: p.MonthlyAddition,
		MaxStocks:       p.MaxStocks,
		CreatedAt:       formatDTO(p.CreatedAt),
		UpdatedAt:       formatDTO(p.UpdatedAt),
	}
}

func newPortfolioDetailResponse(
	p *portfolio.Portfolio,
	holdings []*usecase.HoldingWithValuation,
) *PortfolioDetailResponse {
	items := make([]HoldingDetailResponse, len(holdings))
	for i, hwv := range holdings {
		items[i] = newHoldingDetailResponse(hwv)
	}
	return &PortfolioDetailResponse{
		Portfolio: *newPortfolioResponse(p),
		Holdings:  items,
	}
}

func newHoldingDetailResponse(hwv *usecase.HoldingWithValuation) HoldingDetailResponse {
	resp := HoldingDetailResponse{
		ID:          hwv.Holding.ID,
		Ticker:      hwv.Holding.Ticker,
		AvgBuyPrice: hwv.Holding.AvgBuyPrice,
		Lots:        hwv.Holding.Lots,
	}
	if hwv.StockData != nil {
		resp.CurrentPrice = &hwv.StockData.Price
	}
	if hwv.Valuation != nil {
		resp.GrahamNumber = &hwv.Valuation.GrahamNumber
		resp.EntryPrice = &hwv.Valuation.EntryPrice
		resp.ExitTarget = &hwv.Valuation.ExitTarget
		resp.MarginOfSafety = &hwv.Valuation.MarginOfSafety
		verdict := string(hwv.Valuation.Verdict)
		resp.Verdict = &verdict
	}
	return resp
}

func formatDTO(t time.Time) string {
	return t.UTC().Format(timeLayout)
}
