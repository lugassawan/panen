package checklist

import (
	"fmt"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

// EvaluateInput contains all data needed for auto-check evaluation.
type EvaluateInput struct {
	Action      ActionType
	StockData   *stock.Data
	Valuation   *valuation.ValuationResult
	Holding     *portfolio.Holding // nil for new buy
	Portfolio   *portfolio.Portfolio
	AllHoldings []*portfolio.Holding
	Thresholds  Thresholds
	BuyFeePct   float64
	SellFeePct  float64
	SellTaxPct  float64
}

// CheckResult holds the evaluated result of a single check.
type CheckResult struct {
	Key    string
	Label  string
	Type   CheckType
	Status CheckStatus
	Detail string
}

// EvaluateAutoChecks evaluates all auto-checks for the given input.
func EvaluateAutoChecks(input EvaluateInput) []CheckResult {
	defs := AutoCheckDefs(input.Action)
	results := make([]CheckResult, 0, len(defs))

	for _, def := range defs {
		var cr CheckResult
		switch def.Key {
		case "roe_above_min":
			cr = checkROE(input.StockData, input.Thresholds)
		case "der_below_max":
			cr = checkDER(input.StockData, input.Thresholds)
		case "price_below_entry":
			cr = checkPriceBelowEntry(input.StockData, input.Valuation)
		case "position_weight":
			cr = checkPositionWeight(input.StockData, input.Portfolio, input.AllHoldings, input.Thresholds)
		case "current_loss":
			cr = checkCurrentLoss(input.StockData, input.Holding)
		case "new_avg_price":
			cr = checkNewAvgPrice(input.StockData, input.Holding)
		case "dy_above_min":
			cr = checkDYAboveMin(input.StockData, input.Thresholds)
		case "payout_sustainable":
			cr = checkPayoutSustainable(input.StockData, input.Thresholds)
		case "price_above_exit":
			cr = checkPriceAboveExit(input.StockData, input.Valuation)
		case "capital_gain":
			cr = checkCapitalGain(input.StockData, input.Holding, input.BuyFeePct, input.SellFeePct, input.SellTaxPct)
		case "fundamentals_stable":
			cr = checkFundamentalsStable(input.StockData, input.Thresholds)
		case "dividend_maintained":
			cr = checkDividendMaintained(input.StockData)
		}
		cr.Key = def.Key
		cr.Label = def.Label
		cr.Type = def.Type
		results = append(results, cr)
	}

	return results
}

func checkROE(data *stock.Data, th Thresholds) CheckResult {
	status := CheckStatusFail
	if data.ROE >= th.MinROE {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("ROE: %.2f%% (min: %.2f%%)", data.ROE, th.MinROE),
	}
}

func checkDER(data *stock.Data, th Thresholds) CheckResult {
	status := CheckStatusFail
	if data.DER <= th.MaxDER {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("DER: %.2f (max: %.2f)", data.DER, th.MaxDER),
	}
}

func checkPriceBelowEntry(data *stock.Data, val *valuation.ValuationResult) CheckResult {
	status := CheckStatusFail
	if data.Price <= val.EntryPrice {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Price: Rp %.0f (entry: Rp %.0f)", data.Price, val.EntryPrice),
	}
}

func checkPositionWeight(
	data *stock.Data,
	_ *portfolio.Portfolio,
	allHoldings []*portfolio.Holding,
	th Thresholds,
) CheckResult {
	var totalValue float64
	var currentValue float64

	for _, h := range allHoldings {
		value := h.AvgBuyPrice * float64(h.Lots) * 100
		if h.Ticker == data.Ticker {
			value = data.Price * float64(h.Lots) * 100
			currentValue = value
		}
		totalValue += value
	}

	var weight float64
	if totalValue > 0 {
		weight = (currentValue / totalValue) * 100
	}

	status := CheckStatusFail
	if weight <= th.MaxPositionPct {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Weight: %.2f%% (max: %.2f%%)", weight, th.MaxPositionPct),
	}
}

func checkCurrentLoss(data *stock.Data, h *portfolio.Holding) CheckResult {
	lossPct := ((data.Price - h.AvgBuyPrice) / h.AvgBuyPrice) * 100

	status := CheckStatusFail
	if data.Price < h.AvgBuyPrice {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Current P&L: %.2f%%", lossPct),
	}
}

func checkNewAvgPrice(data *stock.Data, h *portfolio.Holding) CheckResult {
	status := CheckStatusFail
	if data.Price < h.AvgBuyPrice {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Current avg: Rp %.0f, buying at: Rp %.0f", h.AvgBuyPrice, data.Price),
	}
}

func checkDYAboveMin(data *stock.Data, th Thresholds) CheckResult {
	status := CheckStatusFail
	if data.DividendYield >= th.MinDY {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("DY: %.2f%% (min: %.2f%%)", data.DividendYield, th.MinDY),
	}
}

func checkPayoutSustainable(data *stock.Data, th Thresholds) CheckResult {
	status := CheckStatusFail
	if data.PayoutRatio <= th.MaxPayoutRatio {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Payout ratio: %.2f%% (max: %.2f%%)", data.PayoutRatio, th.MaxPayoutRatio),
	}
}

func checkPriceAboveExit(data *stock.Data, val *valuation.ValuationResult) CheckResult {
	status := CheckStatusFail
	if data.Price >= val.ExitTarget {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Price: Rp %.0f (target: Rp %.0f)", data.Price, val.ExitTarget),
	}
}

func checkCapitalGain(data *stock.Data, h *portfolio.Holding, buyFeePct, sellFeePct, sellTaxPct float64) CheckResult {
	shares := float64(h.Lots) * 100
	buyCost := h.AvgBuyPrice * shares * (1 + buyFeePct/100)
	sellProceeds := data.Price * shares * (1 - sellFeePct/100 - sellTaxPct/100)
	gainPct := ((sellProceeds - buyCost) / buyCost) * 100

	status := CheckStatusFail
	if gainPct > 0 {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Capital gain: %.2f%%", gainPct),
	}
}

func checkFundamentalsStable(data *stock.Data, th Thresholds) CheckResult {
	roeOK := data.ROE >= th.MinROE
	derOK := data.DER <= th.MaxDER

	status := CheckStatusFail
	if roeOK && derOK {
		status = CheckStatusPass
	}

	var detail string
	switch {
	case roeOK && derOK:
		detail = fmt.Sprintf("ROE: %.2f%% (pass), DER: %.2f (pass)", data.ROE, data.DER)
	case roeOK && !derOK:
		detail = fmt.Sprintf("ROE: %.2f%% (pass), DER: %.2f (fail, max: %.2f)", data.ROE, data.DER, th.MaxDER)
	case !roeOK && derOK:
		detail = fmt.Sprintf("ROE: %.2f%% (fail, min: %.2f%%), DER: %.2f (pass)", data.ROE, th.MinROE, data.DER)
	default:
		detail = fmt.Sprintf(
			"ROE: %.2f%% (fail, min: %.2f%%), DER: %.2f (fail, max: %.2f)",
			data.ROE,
			th.MinROE,
			data.DER,
			th.MaxDER,
		)
	}

	return CheckResult{
		Status: status,
		Detail: detail,
	}
}

func checkDividendMaintained(data *stock.Data) CheckResult {
	status := CheckStatusFail
	if data.DividendYield > 0 {
		status = CheckStatusPass
	}
	return CheckResult{
		Status: status,
		Detail: fmt.Sprintf("Dividend yield: %.2f%%", data.DividendYield),
	}
}
