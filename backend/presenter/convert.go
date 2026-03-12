package presenter

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/alert"
	"github.com/lugassawan/panen/backend/domain/brokerage"
	"github.com/lugassawan/panen/backend/domain/checklist"
	"github.com/lugassawan/panen/backend/domain/dashboard"
	"github.com/lugassawan/panen/backend/domain/dividend"
	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/trailingstop"
	"github.com/lugassawan/panen/backend/domain/transaction"
	"github.com/lugassawan/panen/backend/domain/valuation"
	"github.com/lugassawan/panen/backend/domain/watchlist"
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
		PBVBand:        newBandStatsResponse(result.PBVBand),
		PERBand:        newBandStatsResponse(result.PERBand),
		Verdict:        string(result.Verdict),
		RiskProfile:    riskProfile,
		FetchedAt:      formatDTO(data.FetchedAt),
		Source:         data.Source,
	}
}

func newBandStatsResponse(band *valuation.BandStats) *BandStatsResponse {
	if band == nil {
		return nil
	}
	return &BandStatsResponse{
		Min:    band.Min,
		Max:    band.Max,
		Avg:    band.Avg,
		Median: band.Median,
	}
}

func newBrokerageAccountResponse(acct *brokerage.Account) *BrokerageAccountResponse {
	return &BrokerageAccountResponse{
		ID:          acct.ID,
		BrokerName:  acct.BrokerName,
		BrokerCode:  acct.BrokerCode,
		BuyFeePct:   acct.BuyFeePct,
		SellFeePct:  acct.SellFeePct,
		SellTaxPct:  acct.SellTaxPct,
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
	if hwv.TrailingStop != nil {
		resp.TrailingStop = newTrailingStopResponse(hwv.TrailingStop)
	}
	if hwv.DividendMetrics != nil {
		resp.DividendMetrics = newDividendMetricsResponse(hwv.DividendMetrics)
	}
	return resp
}

func newTrailingStopResponse(ts *trailingstop.TrailingStopResult) *TrailingStopResponse {
	exits := make([]FundamentalExitResponse, len(ts.FundamentalExits))
	for i, fe := range ts.FundamentalExits {
		exits[i] = FundamentalExitResponse{
			Key:       fe.Key,
			Label:     fe.Label,
			Detail:    fe.Detail,
			Triggered: fe.Triggered,
		}
	}
	return &TrailingStopResponse{
		PeakPrice:        ts.PeakPrice,
		StopPercentage:   ts.StopPct,
		StopPrice:        ts.StopPrice,
		Triggered:        ts.Triggered,
		FundamentalExits: exits,
	}
}

func newDividendMetricsResponse(m *dividend.DividendMetrics) *DividendMetricsResponse {
	return &DividendMetricsResponse{
		Indicator:      string(m.Indicator),
		AnnualDPS:      m.AnnualDPS,
		YieldOnCost:    m.YieldOnCost,
		ProjectedYoC:   m.ProjectedYoC,
		PortfolioYield: m.PortfolioYield,
	}
}

func newDividendRankItemResponse(item dividend.RankItem) DividendRankItemResponse {
	return DividendRankItemResponse{
		Ticker:      item.Ticker,
		Indicator:   string(item.Indicator),
		DY:          item.DY,
		YoC:         item.YieldOnCost,
		PayoutRatio: item.PayoutRatio,
		PositionPct: item.PositionPct,
		Score:       item.Score,
		IsHolding:   item.IsHolding,
	}
}

func newWatchlistResponse(w *watchlist.Watchlist) *WatchlistResponse {
	return &WatchlistResponse{
		ID:        w.ID,
		Name:      w.Name,
		CreatedAt: formatDTO(w.CreatedAt),
		UpdatedAt: formatDTO(w.UpdatedAt),
	}
}

func newWatchlistItemResponse(item *usecase.WatchlistItemWithData) *WatchlistItemResponse {
	resp := &WatchlistItemResponse{
		Ticker: item.Ticker,
		Sector: item.Sector,
	}
	if item.StockData != nil {
		resp.Price = &item.StockData.Price
		resp.ROE = &item.StockData.ROE
		resp.DER = &item.StockData.DER
		resp.EPS = &item.StockData.EPS
		resp.DividendYield = &item.StockData.DividendYield
		resp.PayoutRatio = &item.StockData.PayoutRatio
		fetchedAt := formatDTO(item.StockData.FetchedAt)
		resp.FetchedAt = &fetchedAt
	}
	if item.Valuation != nil {
		resp.GrahamNumber = &item.Valuation.GrahamNumber
		resp.EntryPrice = &item.Valuation.EntryPrice
		resp.ExitTarget = &item.Valuation.ExitTarget
		verdict := string(item.Valuation.Verdict)
		resp.Verdict = &verdict
	}
	return resp
}

func newCheckResultResponse(cr checklist.CheckResult) CheckResultResponse {
	return CheckResultResponse{
		Key:    cr.Key,
		Label:  cr.Label,
		Type:   string(cr.Type),
		Status: string(cr.Status),
		Detail: cr.Detail,
	}
}

func newSuggestionResponse(s *checklist.Suggestion) *SuggestionResponse {
	if s == nil {
		return nil
	}
	return &SuggestionResponse{
		Action:          string(s.Action),
		Ticker:          s.Ticker,
		Lots:            s.Lots,
		PricePerShare:   s.PricePerShare,
		GrossCost:       s.GrossCost,
		Fee:             s.Fee,
		Tax:             s.Tax,
		NetCost:         s.NetCost,
		NewAvgBuyPrice:  s.NewAvgBuyPrice,
		NewPositionLots: s.NewPositionLots,
		NewPositionPct:  s.NewPositionPct,
		CapitalGainPct:  s.CapitalGainPct,
	}
}

func newChecklistEvaluationResponse(eval *usecase.ChecklistEvaluation) *ChecklistEvaluationResponse {
	checks := make([]CheckResultResponse, len(eval.Checks))
	for i, cr := range eval.Checks {
		checks[i] = newCheckResultResponse(cr)
	}
	return &ChecklistEvaluationResponse{
		Action:     string(eval.Action),
		Ticker:     eval.Ticker,
		Checks:     checks,
		AllPassed:  eval.AllPassed,
		Suggestion: newSuggestionResponse(eval.Suggestion),
	}
}

func newScreenerItemResponse(r *usecase.ScreenResult) *ScreenerItemResponse {
	resp := &ScreenerItemResponse{
		Ticker: r.Ticker,
		Sector: r.Sector,
		Passed: r.Passed,
		Score:  r.Score,
	}
	if r.StockData != nil {
		resp.Price = &r.StockData.Price
		resp.ROE = &r.StockData.ROE
		resp.DER = &r.StockData.DER
		resp.EPS = &r.StockData.EPS
		resp.PBV = &r.StockData.PBV
		resp.PER = &r.StockData.PER
		resp.DividendYield = &r.StockData.DividendYield
		fetchedAt := formatDTO(r.StockData.FetchedAt)
		resp.FetchedAt = &fetchedAt
	}
	if r.Valuation != nil {
		resp.GrahamNumber = &r.Valuation.GrahamNumber
		resp.EntryPrice = &r.Valuation.EntryPrice
		resp.ExitTarget = &r.Valuation.ExitTarget
		verdict := string(r.Valuation.Verdict)
		resp.Verdict = &verdict
	}
	checks := make([]ScreenerCheckResponse, len(r.Checks))
	for i, c := range r.Checks {
		checks[i] = ScreenerCheckResponse{
			Key:    c.Key,
			Label:  c.Label,
			Status: string(c.Status),
			Value:  c.Value,
			Limit:  c.Limit,
		}
	}
	resp.Checks = checks
	return resp
}

func formatDTO(t time.Time) string {
	return t.UTC().Format(timeLayout)
}

func newMonthlyPaydayResponse(status *usecase.MonthlyPaydayStatus) *MonthlyPaydayResponse {
	portfolios := make([]PortfolioPaydayItemResponse, len(status.Portfolios))
	for i, ps := range status.Portfolios {
		portfolios[i] = newPortfolioPaydayItemResponse(ps)
	}
	return &MonthlyPaydayResponse{
		Month:         status.Month,
		PaydayDay:     status.PaydayDay,
		Portfolios:    portfolios,
		TotalExpected: status.TotalExpected,
	}
}

func newPortfolioPaydayItemResponse(ps usecase.PortfolioPaydayStatus) PortfolioPaydayItemResponse {
	resp := PortfolioPaydayItemResponse{
		PortfolioID:   ps.PortfolioID,
		PortfolioName: ps.PortfolioName,
		Mode:          ps.Mode,
		Expected:      ps.Expected,
		Actual:        ps.Actual,
		Status:        ps.Status,
	}
	if ps.DeferUntil != nil {
		s := formatDTO(*ps.DeferUntil)
		resp.DeferUntil = &s
	}
	return resp
}

func newCashFlowSummaryResponse(summary *usecase.CashFlowSummary) *CashFlowSummaryResponse {
	items := make([]CashFlowItemResponse, len(summary.Items))
	for i, item := range summary.Items {
		items[i] = newCashFlowItemResponse(item)
	}
	return &CashFlowSummaryResponse{
		Items:         items,
		TotalInflow:   summary.TotalInflow,
		TotalDeployed: summary.TotalDeployed,
		Balance:       summary.Balance,
	}
}

func newCashFlowItemResponse(item usecase.CashFlowItem) CashFlowItemResponse {
	return CashFlowItemResponse{
		ID:          item.ID,
		PortfolioID: item.PortfolioID,
		Type:        item.Type,
		Amount:      item.Amount,
		Date:        formatDTO(item.Date),
		Note:        item.Note,
		CreatedAt:   formatDTO(item.CreatedAt),
	}
}

func newTransactionRecordResponse(r transaction.Record) TransactionRecordResponse {
	return TransactionRecordResponse{
		ID:            r.ID,
		Type:          string(r.Type),
		Date:          formatDTO(r.Date),
		Ticker:        r.Ticker,
		PortfolioID:   r.PortfolioID,
		PortfolioName: r.PortfolioName,
		Lots:          r.Lots,
		Price:         r.Price,
		Fee:           r.Fee,
		Tax:           r.Tax,
		Total:         r.Total,
		CreatedAt:     formatDTO(r.CreatedAt),
	}
}

func newTransactionSummaryResponse(s *transaction.Summary) TransactionSummaryResponse {
	if s == nil {
		return TransactionSummaryResponse{}
	}
	return TransactionSummaryResponse{
		TotalBuyAmount:      s.TotalBuyAmount,
		TotalSellAmount:     s.TotalSellAmount,
		TotalDividendAmount: s.TotalDividendAmount,
		TotalFees:           s.TotalFees,
		TransactionCount:    s.TransactionCount,
	}
}

func emptyDashboardOverview() *DashboardOverviewResponse {
	return &DashboardOverviewResponse{
		Portfolios:          []PortfolioSummaryResponse{},
		TopGainers:          []HoldingPLResponse{},
		TopLosers:           []HoldingPLResponse{},
		PortfolioAllocation: []AllocationItemResponse{},
		SectorAllocation:    []AllocationItemResponse{},
		RecentTransactions:  []TransactionRecordResponse{},
	}
}

func newDashboardOverviewResponse(o *dashboard.Overview) *DashboardOverviewResponse {
	portfolios := make([]PortfolioSummaryResponse, len(o.Portfolios))
	for i, ps := range o.Portfolios {
		portfolios[i] = PortfolioSummaryResponse{
			ID: ps.ID, Name: ps.Name, Mode: ps.Mode,
			MarketValue: ps.MarketValue, CostBasis: ps.CostBasis,
			PLAmount: ps.PLAmount, PLPercent: ps.PLPercent, Weight: ps.Weight,
		}
	}

	gainers := make([]HoldingPLResponse, len(o.TopGainers))
	for i, h := range o.TopGainers {
		gainers[i] = newHoldingPLResponse(h)
	}
	losers := make([]HoldingPLResponse, len(o.TopLosers))
	for i, h := range o.TopLosers {
		losers[i] = newHoldingPLResponse(h)
	}

	portfolioAlloc := make([]AllocationItemResponse, len(o.PortfolioAllocation))
	for i, a := range o.PortfolioAllocation {
		portfolioAlloc[i] = AllocationItemResponse{Label: a.Label, Value: a.Value, Pct: a.Pct}
	}
	sectorAlloc := make([]AllocationItemResponse, len(o.SectorAllocation))
	for i, a := range o.SectorAllocation {
		sectorAlloc[i] = AllocationItemResponse{Label: a.Label, Value: a.Value, Pct: a.Pct}
	}

	txns := make([]TransactionRecordResponse, len(o.RecentTransactions))
	for i, r := range o.RecentTransactions {
		txns[i] = newTransactionRecordResponse(r)
	}

	return &DashboardOverviewResponse{
		TotalMarketValue:    o.TotalMarketValue,
		TotalCostBasis:      o.TotalCostBasis,
		TotalPLAmount:       o.TotalPLAmount,
		TotalPLPercent:      o.TotalPLPercent,
		TotalDividendIncome: o.TotalDividendIncome,
		Portfolios:          portfolios,
		TopGainers:          gainers,
		TopLosers:           losers,
		PortfolioAllocation: portfolioAlloc,
		SectorAllocation:    sectorAlloc,
		RecentTransactions:  txns,
		WinRate:             o.WinRate,
		HoldingCount:        o.HoldingCount,
		WinningCount:        o.WinningCount,
	}
}

func newHoldingPLResponse(h dashboard.HoldingPL) HoldingPLResponse {
	return HoldingPLResponse{
		Ticker: h.Ticker, PortfolioID: h.PortfolioID, PortfolioName: h.PortfolioName,
		MarketValue: h.MarketValue, CostBasis: h.CostBasis,
		PLAmount: h.PLAmount, PLPercent: h.PLPercent,
	}
}

func newFundamentalAlertResponse(a *alert.FundamentalAlert) FundamentalAlertResponse {
	resp := FundamentalAlertResponse{
		ID:         a.ID,
		Ticker:     a.Ticker,
		Metric:     a.Metric,
		Severity:   string(a.Severity),
		OldValue:   a.OldValue,
		NewValue:   a.NewValue,
		ChangePct:  a.ChangePct,
		Status:     string(a.Status),
		DetectedAt: formatDTO(a.DetectedAt),
	}
	if a.ResolvedAt != nil {
		resp.ResolvedAt = formatDTO(*a.ResolvedAt)
	}
	return resp
}
