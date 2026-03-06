package checklist

import (
	"fmt"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

// EvaluateInput contains all data needed for auto-check evaluation.
type EvaluateInput struct {
	Action           ActionType
	StockData        *stock.Data
	Valuation        *valuation.ValuationResult
	Holding          *portfolio.Holding // nil for new buy
	Portfolio        *portfolio.Portfolio
	AllHoldings      []*portfolio.Holding
	Thresholds       Thresholds
	BuyFeePct        float64
	SellFeePct       float64
	SellTaxPct       float64
	HasCriticalAlert bool
}

// CheckResult holds the evaluated result of a single check.
type CheckResult struct {
	Key    string
	Label  string
	Type   CheckType
	Status CheckStatus
	Detail string
}

type checkFunc func(EvaluateInput) CheckResult

var checkRegistry = map[string]checkFunc{
	"roe_above_min":     func(in EvaluateInput) CheckResult { return checkROE(in.StockData, in.Thresholds) },
	"der_below_max":     func(in EvaluateInput) CheckResult { return checkDER(in.StockData, in.Thresholds) },
	"price_below_entry": func(in EvaluateInput) CheckResult { return checkPriceBelowEntry(in.StockData, in.Valuation) },
	"position_weight": func(in EvaluateInput) CheckResult {
		return checkPositionWeight(in.StockData, in.AllHoldings, in.Thresholds)
	},
	"current_loss":  func(in EvaluateInput) CheckResult { return checkCurrentLoss(in.StockData, in.Holding) },
	"new_avg_price": func(in EvaluateInput) CheckResult { return checkNewAvgPrice(in.StockData, in.Holding) },
	"dy_above_min":  func(in EvaluateInput) CheckResult { return checkDYAboveMin(in.StockData, in.Thresholds) },
	"payout_sustainable": func(in EvaluateInput) CheckResult {
		return checkPayoutSustainable(in.StockData, in.Thresholds)
	},
	"price_above_exit": func(in EvaluateInput) CheckResult { return checkPriceAboveExit(in.StockData, in.Valuation) },
	"capital_gain": func(in EvaluateInput) CheckResult {
		return checkCapitalGain(in.StockData, in.Holding, in.BuyFeePct, in.SellFeePct, in.SellTaxPct)
	},
	"fundamentals_stable": func(in EvaluateInput) CheckResult {
		return checkFundamentalsStable(in.StockData, in.Thresholds)
	},
	"dividend_maintained": func(in EvaluateInput) CheckResult { return checkDividendMaintained(in.StockData) },
	"no_critical_alerts":  func(in EvaluateInput) CheckResult { return checkNoCriticalAlerts(in.HasCriticalAlert) },
}

// EvaluateAutoChecks evaluates all auto-checks for the given input.
func EvaluateAutoChecks(input EvaluateInput) []CheckResult {
	defs := AutoCheckDefs(input.Action)
	results := make([]CheckResult, 0, len(defs))

	for _, def := range defs {
		fn, ok := checkRegistry[def.Key]
		if !ok {
			continue
		}
		cr := fn(input)
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

func checkNoCriticalAlerts(hasCritical bool) CheckResult {
	status := CheckStatusPass
	detail := "No critical fundamental alerts"
	if hasCritical {
		status = CheckStatusFail
		detail = "Critical fundamental alert active for this ticker"
	}
	return CheckResult{
		Status: status,
		Detail: detail,
	}
}
