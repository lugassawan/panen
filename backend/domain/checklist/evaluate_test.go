package checklist

import (
	"strings"
	"testing"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
	"github.com/lugassawan/panen/backend/domain/valuation"
)

func TestCheckROE(t *testing.T) {
	th := Thresholds{MinROE: 12}

	tests := []struct {
		name       string
		roe        float64
		wantStatus CheckStatus
	}{
		{name: "above threshold", roe: 15, wantStatus: CheckStatusPass},
		{name: "at threshold", roe: 12, wantStatus: CheckStatusPass},
		{name: "below threshold", roe: 10, wantStatus: CheckStatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{ROE: tt.roe}
			got := checkROE(data, th)
			if got.Status != tt.wantStatus {
				t.Errorf("checkROE(ROE=%.2f) status = %q, want %q", tt.roe, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkROE() detail is empty")
			}
		})
	}
}

func TestCheckDER(t *testing.T) {
	th := Thresholds{MaxDER: 1.0}

	tests := []struct {
		name       string
		der        float64
		wantStatus CheckStatus
	}{
		{name: "below threshold", der: 0.5, wantStatus: CheckStatusPass},
		{name: "at threshold", der: 1.0, wantStatus: CheckStatusPass},
		{name: "above threshold", der: 1.5, wantStatus: CheckStatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{DER: tt.der}
			got := checkDER(data, th)
			if got.Status != tt.wantStatus {
				t.Errorf("checkDER(DER=%.2f) status = %q, want %q", tt.der, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkDER() detail is empty")
			}
		})
	}
}

func TestCheckPriceBelowEntry(t *testing.T) {
	tests := []struct {
		name           string
		price          float64
		entry          float64
		wantStatus     CheckStatus
		wantDetailSubs []string
	}{
		{
			name: "below entry", price: 900, entry: 1000,
			wantStatus:     CheckStatusPass,
			wantDetailSubs: []string{"Rp 900", "entry: Rp 1000"},
		},
		{
			name: "at entry", price: 1000, entry: 1000,
			wantStatus:     CheckStatusPass,
			wantDetailSubs: []string{"Rp 1000", "entry: Rp 1000"},
		},
		{
			name: "above entry", price: 1100, entry: 1000,
			wantStatus:     CheckStatusFail,
			wantDetailSubs: []string{"Rp 1100", "entry: Rp 1000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{Price: tt.price}
			val := &valuation.ValuationResult{EntryPrice: tt.entry}
			got := checkPriceBelowEntry(data, val)
			if got.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", got.Status, tt.wantStatus)
			}
			for _, sub := range tt.wantDetailSubs {
				if !strings.Contains(got.Detail, sub) {
					t.Errorf("detail %q missing substring %q", got.Detail, sub)
				}
			}
		})
	}
}

func TestCheckPositionWeight(t *testing.T) {
	th := Thresholds{MaxPositionPct: 20}

	tests := []struct {
		name        string
		ticker      string
		price       float64
		allHoldings []*portfolio.Holding
		wantStatus  CheckStatus
	}{
		{
			name:   "within limit",
			ticker: "BBCA",
			price:  1000,
			allHoldings: []*portfolio.Holding{
				{Ticker: "BBCA", AvgBuyPrice: 900, Lots: 10},
				{Ticker: "BBRI", AvgBuyPrice: 500, Lots: 100},
			},
			wantStatus: CheckStatusPass,
		},
		{
			name:   "exceeds limit",
			ticker: "BBCA",
			price:  1000,
			allHoldings: []*portfolio.Holding{
				{Ticker: "BBCA", AvgBuyPrice: 900, Lots: 100},
				{Ticker: "BBRI", AvgBuyPrice: 500, Lots: 10},
			},
			wantStatus: CheckStatusFail,
		},
		{
			name:   "new buy no holding",
			ticker: "BBCA",
			price:  1000,
			allHoldings: []*portfolio.Holding{
				{Ticker: "BBRI", AvgBuyPrice: 500, Lots: 10},
			},
			wantStatus: CheckStatusPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{Ticker: tt.ticker, Price: tt.price}
			got := checkPositionWeight(data, tt.allHoldings, th)
			if got.Status != tt.wantStatus {
				t.Errorf("checkPositionWeight(%s) status = %q, want %q", tt.name, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkPositionWeight() detail is empty")
			}
		})
	}
}

func TestCheckCurrentLoss(t *testing.T) {
	tests := []struct {
		name            string
		price           float64
		avgBuy          float64
		wantStatus      CheckStatus
		wantDetailContP bool // whether detail should contain negative P&L
	}{
		{name: "at loss", price: 800, avgBuy: 1000, wantStatus: CheckStatusPass, wantDetailContP: true},
		{name: "at profit", price: 1200, avgBuy: 1000, wantStatus: CheckStatusFail, wantDetailContP: false},
		{name: "break even", price: 1000, avgBuy: 1000, wantStatus: CheckStatusFail, wantDetailContP: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{Price: tt.price}
			h := &portfolio.Holding{AvgBuyPrice: tt.avgBuy}
			got := checkCurrentLoss(data, h)
			if got.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q (detail: %s)", got.Status, tt.wantStatus, got.Detail)
			}
			if !strings.Contains(got.Detail, "P&L") {
				t.Errorf("detail %q missing P&L label", got.Detail)
			}
			hasNegative := strings.Contains(got.Detail, "-")
			if tt.wantDetailContP && !hasNegative {
				t.Errorf("detail %q should contain negative percentage", got.Detail)
			}
		})
	}
}

func TestCheckNewAvgPrice(t *testing.T) {
	tests := []struct {
		name       string
		price      float64
		avgBuy     float64
		wantStatus CheckStatus
		wantAvgSub string
		wantBuySub string
	}{
		{
			name: "lower price", price: 800, avgBuy: 1000,
			wantStatus: CheckStatusPass, wantAvgSub: "Rp 1000", wantBuySub: "Rp 800",
		},
		{
			name: "higher price", price: 1200, avgBuy: 1000,
			wantStatus: CheckStatusFail, wantAvgSub: "Rp 1000", wantBuySub: "Rp 1200",
		},
		{
			name: "equal price", price: 1000, avgBuy: 1000,
			wantStatus: CheckStatusFail, wantAvgSub: "Rp 1000", wantBuySub: "buying at: Rp 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{Price: tt.price}
			h := &portfolio.Holding{AvgBuyPrice: tt.avgBuy}
			got := checkNewAvgPrice(data, h)
			if got.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", got.Status, tt.wantStatus)
			}
			if !strings.Contains(got.Detail, tt.wantAvgSub) {
				t.Errorf("detail %q missing avg substring %q", got.Detail, tt.wantAvgSub)
			}
			if !strings.Contains(got.Detail, tt.wantBuySub) {
				t.Errorf("detail %q missing buy substring %q", got.Detail, tt.wantBuySub)
			}
		})
	}
}

func TestCheckDYAboveMin(t *testing.T) {
	th := Thresholds{MinDY: 3}

	tests := []struct {
		name       string
		dy         float64
		wantStatus CheckStatus
	}{
		{name: "above threshold", dy: 5, wantStatus: CheckStatusPass},
		{name: "at threshold", dy: 3, wantStatus: CheckStatusPass},
		{name: "below threshold", dy: 2, wantStatus: CheckStatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{DividendYield: tt.dy}
			got := checkDYAboveMin(data, th)
			if got.Status != tt.wantStatus {
				t.Errorf("checkDYAboveMin(DY=%.2f) status = %q, want %q", tt.dy, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkDYAboveMin() detail is empty")
			}
		})
	}
}

func TestCheckPayoutSustainable(t *testing.T) {
	th := Thresholds{MaxPayoutRatio: 75}

	tests := []struct {
		name       string
		payout     float64
		wantStatus CheckStatus
	}{
		{name: "below threshold", payout: 50, wantStatus: CheckStatusPass},
		{name: "at threshold", payout: 75, wantStatus: CheckStatusPass},
		{name: "above threshold", payout: 90, wantStatus: CheckStatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{PayoutRatio: tt.payout}
			got := checkPayoutSustainable(data, th)
			if got.Status != tt.wantStatus {
				t.Errorf("checkPayoutSustainable(payout=%.2f) status = %q, want %q",
					tt.payout, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkPayoutSustainable() detail is empty")
			}
		})
	}
}

func TestCheckPriceAboveExit(t *testing.T) {
	tests := []struct {
		name       string
		price      float64
		exit       float64
		wantStatus CheckStatus
	}{
		{name: "above target", price: 1200, exit: 1000, wantStatus: CheckStatusPass},
		{name: "at target", price: 1000, exit: 1000, wantStatus: CheckStatusPass},
		{name: "below target", price: 800, exit: 1000, wantStatus: CheckStatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{Price: tt.price}
			val := &valuation.ValuationResult{ExitTarget: tt.exit}
			got := checkPriceAboveExit(data, val)
			if got.Status != tt.wantStatus {
				t.Errorf("checkPriceAboveExit(price=%.0f, exit=%.0f) status = %q, want %q",
					tt.price, tt.exit, got.Status, tt.wantStatus)
			}
			if !strings.Contains(got.Detail, "target") {
				t.Errorf("detail %q missing 'target' label", got.Detail)
			}
		})
	}
}

func TestCheckCapitalGain(t *testing.T) {
	tests := []struct {
		name       string
		price      float64
		avgBuy     float64
		lots       int
		buyFee     float64
		sellFee    float64
		sellTax    float64
		wantStatus CheckStatus
	}{
		{
			name:       "with gain",
			price:      1500,
			avgBuy:     1000,
			lots:       10,
			buyFee:     0.15,
			sellFee:    0.15,
			sellTax:    0.10,
			wantStatus: CheckStatusPass,
		},
		{
			name:       "with loss",
			price:      800,
			avgBuy:     1000,
			lots:       10,
			buyFee:     0.15,
			sellFee:    0.15,
			sellTax:    0.10,
			wantStatus: CheckStatusFail,
		},
		{
			name:       "break even before fees results in loss",
			price:      1000,
			avgBuy:     1000,
			lots:       10,
			buyFee:     0.15,
			sellFee:    0.15,
			sellTax:    0.10,
			wantStatus: CheckStatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{Price: tt.price}
			h := &portfolio.Holding{AvgBuyPrice: tt.avgBuy, Lots: tt.lots}
			got := checkCapitalGain(data, h, tt.buyFee, tt.sellFee, tt.sellTax)
			if got.Status != tt.wantStatus {
				t.Errorf("checkCapitalGain(%s) status = %q, want %q", tt.name, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkCapitalGain() detail is empty")
			}
		})
	}
}

func TestCheckFundamentalsStable(t *testing.T) {
	th := Thresholds{MinROE: 12, MaxDER: 1.0}

	tests := []struct {
		name       string
		roe        float64
		der        float64
		wantStatus CheckStatus
	}{
		{name: "both pass", roe: 15, der: 0.8, wantStatus: CheckStatusPass},
		{name: "ROE fails", roe: 8, der: 0.8, wantStatus: CheckStatusFail},
		{name: "DER fails", roe: 15, der: 1.5, wantStatus: CheckStatusFail},
		{name: "both fail", roe: 8, der: 1.5, wantStatus: CheckStatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{ROE: tt.roe, DER: tt.der}
			got := checkFundamentalsStable(data, th)
			if got.Status != tt.wantStatus {
				t.Errorf("checkFundamentalsStable(ROE=%.2f, DER=%.2f) status = %q, want %q",
					tt.roe, tt.der, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkFundamentalsStable() detail is empty")
			}
		})
	}
}

func TestCheckDividendMaintained(t *testing.T) {
	tests := []struct {
		name       string
		dy         float64
		wantStatus CheckStatus
	}{
		{name: "has yield", dy: 3.5, wantStatus: CheckStatusPass},
		{name: "no yield", dy: 0, wantStatus: CheckStatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &stock.Data{DividendYield: tt.dy}
			got := checkDividendMaintained(data)
			if got.Status != tt.wantStatus {
				t.Errorf("checkDividendMaintained(DY=%.2f) status = %q, want %q",
					tt.dy, got.Status, tt.wantStatus)
			}
			if got.Detail == "" {
				t.Error("checkDividendMaintained() detail is empty")
			}
		})
	}
}

func TestEvaluateAutoChecks(t *testing.T) {
	baseData := &stock.Data{
		Ticker:        "BBCA",
		Price:         900,
		ROE:           15,
		DER:           0.5,
		DividendYield: 4,
		PayoutRatio:   50,
	}
	baseVal := &valuation.ValuationResult{
		EntryPrice: 1000,
		ExitTarget: 1200,
	}
	baseTh := Thresholds{
		MinROE:         12,
		MaxDER:         1.0,
		MaxPositionPct: 20,
		MinDY:          3,
		MaxPayoutRatio: 75,
	}
	baseHolding := &portfolio.Holding{
		Ticker:      "BBCA",
		AvgBuyPrice: 1000,
		Lots:        10,
	}

	tests := []struct {
		name      string
		action    ActionType
		wantCount int
	}{
		{name: "BUY", action: ActionBuy, wantCount: 5},
		{name: "AVERAGE_DOWN", action: ActionAverageDown, wantCount: 6},
		{name: "AVERAGE_UP", action: ActionAverageUp, wantCount: 5},
		{name: "SELL_EXIT", action: ActionSellExit, wantCount: 2},
		{name: "SELL_STOP", action: ActionSellStop, wantCount: 2},
		{name: "HOLD", action: ActionHold, wantCount: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := EvaluateInput{
				Action:    tt.action,
				StockData: baseData,
				Valuation: baseVal,
				Holding:   baseHolding,
				Portfolio: &portfolio.Portfolio{},
				AllHoldings: []*portfolio.Holding{
					baseHolding,
				},
				Thresholds: baseTh,
				BuyFeePct:  0.15,
				SellFeePct: 0.15,
				SellTaxPct: 0.10,
			}

			results := EvaluateAutoChecks(input)
			if len(results) != tt.wantCount {
				t.Errorf("EvaluateAutoChecks(%s) returned %d results, want %d",
					tt.action, len(results), tt.wantCount)
			}

			defs := AutoCheckDefs(tt.action)
			for i, r := range results {
				if r.Key != defs[i].Key {
					t.Errorf("result[%d].Key = %q, want %q", i, r.Key, defs[i].Key)
				}
				if r.Label != defs[i].Label {
					t.Errorf("result[%d].Label = %q, want %q", i, r.Label, defs[i].Label)
				}
				if r.Type != CheckTypeAuto {
					t.Errorf("result[%d].Type = %q, want %q", i, r.Type, CheckTypeAuto)
				}
				if r.Status != CheckStatusPass && r.Status != CheckStatusFail {
					t.Errorf("result[%d].Status = %q, want PASS or FAIL", i, r.Status)
				}
				if r.Detail == "" {
					t.Errorf("result[%d].Detail is empty", i)
				}
			}
		})
	}
}

func TestCheckRegistryCoversAllAutoCheckDefs(t *testing.T) {
	// Collect all unique keys referenced in AutoCheckDefs across all action types.
	actions := []ActionType{
		ActionBuy,
		ActionAverageDown,
		ActionAverageUp,
		ActionSellExit,
		ActionSellStop,
		ActionHold,
	}

	seen := map[string]bool{}
	for _, action := range actions {
		for _, def := range AutoCheckDefs(action) {
			seen[def.Key] = true
		}
	}

	for key := range seen {
		t.Run(key, func(t *testing.T) {
			if _, ok := checkRegistry[key]; !ok {
				t.Errorf("AutoCheckDefs key %q has no registered handler in checkRegistry", key)
			}
		})
	}
}

func TestEvaluateAutoChecksDispatch(t *testing.T) {
	// Verify BUY action dispatches to correct check functions by checking specific results
	input := EvaluateInput{
		Action: ActionBuy,
		StockData: &stock.Data{
			Ticker: "BBCA",
			Price:  900,
			ROE:    15,
			DER:    0.5,
		},
		Valuation: &valuation.ValuationResult{
			EntryPrice: 1000,
		},
		Portfolio: &portfolio.Portfolio{},
		AllHoldings: []*portfolio.Holding{
			{Ticker: "BBCA", AvgBuyPrice: 900, Lots: 10},
			{Ticker: "BBRI", AvgBuyPrice: 500, Lots: 10},
		},
		Thresholds: Thresholds{
			MinROE:         12,
			MaxDER:         1.0,
			MaxPositionPct: 80,
		},
	}

	results := EvaluateAutoChecks(input)

	// All checks should pass with these inputs
	for _, r := range results {
		if r.Status != CheckStatusPass {
			t.Errorf("EvaluateAutoChecks BUY: check %q status = %q, want PASS (detail: %s)",
				r.Key, r.Status, r.Detail)
		}
	}
}
