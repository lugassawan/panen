package checklist

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/lugassawan/panen/backend/domain/portfolio"
	"github.com/lugassawan/panen/backend/domain/stock"
)

func TestComputeSuggestionBuyNewPosition(t *testing.T) {
	input := EvaluateInput{
		Action:      ActionBuy,
		StockData:   &stock.Data{Ticker: "BBCA", Price: 1000},
		Portfolio:   &portfolio.Portfolio{Capital: 10000000},
		Thresholds:  Thresholds{MaxPositionPct: 20},
		AllHoldings: []*portfolio.Holding{},
		BuyFeePct:   0.15,
		SellFeePct:  0.15,
		SellTaxPct:  0.10,
	}

	got, err := ComputeSuggestion(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// maxPositionValue = 10_000_000 * 20 / 100 = 2_000_000
	// lots = floor(2_000_000 / (1000 * 100)) = 20
	if got.Action != ActionBuy {
		t.Errorf("Action = %q, want %q", got.Action, ActionBuy)
	}
	if got.Ticker != "BBCA" {
		t.Errorf("Ticker = %q, want %q", got.Ticker, "BBCA")
	}
	if got.Lots != 20 {
		t.Errorf("Lots = %d, want 20", got.Lots)
	}
	if got.PricePerShare != 1000 {
		t.Errorf("PricePerShare = %.2f, want 1000", got.PricePerShare)
	}
	// GrossCost = 1000 * 20 * 100 = 2_000_000
	if got.GrossCost != 2000000 {
		t.Errorf("GrossCost = %.2f, want 2000000", got.GrossCost)
	}
	// Fee = 2_000_000 * 0.15 / 100 = 3000
	if got.Fee != 3000 {
		t.Errorf("Fee = %.2f, want 3000", got.Fee)
	}
	if got.Tax != 0 {
		t.Errorf("Tax = %.2f, want 0", got.Tax)
	}
	// NetCost = 2_000_000 + 3000 = 2_003_000
	if got.NetCost != 2003000 {
		t.Errorf("NetCost = %.2f, want 2003000", got.NetCost)
	}
	if got.NewAvgBuyPrice != 1000 {
		t.Errorf("NewAvgBuyPrice = %.2f, want 1000", got.NewAvgBuyPrice)
	}
	if got.NewPositionLots != 20 {
		t.Errorf("NewPositionLots = %d, want 20", got.NewPositionLots)
	}
	if got.CapitalGainPct != 0 {
		t.Errorf("CapitalGainPct = %.2f, want 0", got.CapitalGainPct)
	}
	// NewPositionPct: newPositionValue = 1000 * 20 * 100 = 2_000_000
	// totalPortfolioValue = 2_000_000 (only this position)
	// pct = 100%
	if got.NewPositionPct != 100 {
		t.Errorf("NewPositionPct = %.2f, want 100", got.NewPositionPct)
	}
}

func TestComputeSuggestionAverageDown(t *testing.T) {
	input := EvaluateInput{
		Action:     ActionAverageDown,
		StockData:  &stock.Data{Ticker: "BBCA", Price: 800},
		Holding:    &portfolio.Holding{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 10},
		Portfolio:  &portfolio.Portfolio{Capital: 10000000},
		Thresholds: Thresholds{MaxPositionPct: 20},
		AllHoldings: []*portfolio.Holding{
			{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 10},
			{Ticker: "BBRI", AvgBuyPrice: 500, Lots: 20},
		},
		BuyFeePct:  0.15,
		SellFeePct: 0.15,
		SellTaxPct: 0.10,
	}

	got, err := ComputeSuggestion(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// maxPositionValue = 10_000_000 * 20 / 100 = 2_000_000
	// currentPositionValue = 800 * 10 * 100 = 800_000
	// available = 2_000_000 - 800_000 = 1_200_000
	// lots = floor(1_200_000 / (800 * 100)) = 15
	if got.Lots != 15 {
		t.Errorf("Lots = %d, want 15", got.Lots)
	}
	// Weighted avg: existingCost = 1000 * 10 = 10_000
	// newCost = 800 * 15 = 12_000
	// totalLots = 10 + 15 = 25
	// newAvg = (10_000 + 12_000) / 25 = 880
	if got.NewAvgBuyPrice != 880 {
		t.Errorf("NewAvgBuyPrice = %.2f, want 880", got.NewAvgBuyPrice)
	}
	if got.NewPositionLots != 25 {
		t.Errorf("NewPositionLots = %d, want 25", got.NewPositionLots)
	}
	// GrossCost = 800 * 15 * 100 = 1_200_000
	if got.GrossCost != 1200000 {
		t.Errorf("GrossCost = %.2f, want 1200000", got.GrossCost)
	}
	// Fee = 1_200_000 * 0.15 / 100 = 1800
	if got.Fee != 1800 {
		t.Errorf("Fee = %.2f, want 1800", got.Fee)
	}
	if got.Tax != 0 {
		t.Errorf("Tax = %.2f, want 0", got.Tax)
	}
	// NetCost = 1_200_000 + 1800 = 1_201_800
	if got.NetCost != 1201800 {
		t.Errorf("NetCost = %.2f, want 1201800", got.NetCost)
	}
}

func TestComputeSuggestionAverageUp(t *testing.T) {
	input := EvaluateInput{
		Action:     ActionAverageUp,
		StockData:  &stock.Data{Ticker: "BBCA", Price: 1200},
		Holding:    &portfolio.Holding{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 5},
		Portfolio:  &portfolio.Portfolio{Capital: 10000000},
		Thresholds: Thresholds{MaxPositionPct: 20},
		AllHoldings: []*portfolio.Holding{
			{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 5},
		},
		BuyFeePct:  0.15,
		SellFeePct: 0.15,
		SellTaxPct: 0.10,
	}

	got, err := ComputeSuggestion(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// maxPositionValue = 10_000_000 * 20 / 100 = 2_000_000
	// currentPositionValue = 1200 * 5 * 100 = 600_000
	// available = 2_000_000 - 600_000 = 1_400_000
	// lots = floor(1_400_000 / (1200 * 100)) = 11
	if got.Lots != 11 {
		t.Errorf("Lots = %d, want 11", got.Lots)
	}
	if got.Action != ActionAverageUp {
		t.Errorf("Action = %q, want %q", got.Action, ActionAverageUp)
	}
	// existingCost = 1000 * 5 = 5000
	// newCost = 1200 * 11 = 13200
	// totalLots = 16
	// newAvg = 18200 / 16 = 1137.5
	if got.NewAvgBuyPrice != 1137.5 {
		t.Errorf("NewAvgBuyPrice = %.2f, want 1137.50", got.NewAvgBuyPrice)
	}
	if got.NewPositionLots != 16 {
		t.Errorf("NewPositionLots = %d, want 16", got.NewPositionLots)
	}
}

func TestComputeSuggestionSellExit(t *testing.T) {
	input := EvaluateInput{
		Action:    ActionSellExit,
		StockData: &stock.Data{Ticker: "BBCA", Price: 1500},
		Holding:   &portfolio.Holding{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 10},
		Portfolio: &portfolio.Portfolio{Capital: 10000000},
		AllHoldings: []*portfolio.Holding{
			{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 10},
		},
		BuyFeePct:  0.15,
		SellFeePct: 0.15,
		SellTaxPct: 0.10,
	}

	got, err := ComputeSuggestion(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// shares = 10 * 100 = 1000
	// GrossCost = 1500 * 1000 = 1_500_000
	if got.GrossCost != 1500000 {
		t.Errorf("GrossCost = %.2f, want 1500000", got.GrossCost)
	}
	// Fee = 1_500_000 * 0.15 / 100 = 2250
	if got.Fee != 2250 {
		t.Errorf("Fee = %.2f, want 2250", got.Fee)
	}
	// Tax = 1_500_000 * 0.10 / 100 = 1500
	if got.Tax != 1500 {
		t.Errorf("Tax = %.2f, want 1500", got.Tax)
	}
	// NetCost = 1_500_000 - 2250 - 1500 = 1_496_250
	if got.NetCost != 1496250 {
		t.Errorf("NetCost = %.2f, want 1496250", got.NetCost)
	}
	if got.Lots != 10 {
		t.Errorf("Lots = %d, want 10", got.Lots)
	}
	if got.PricePerShare != 1500 {
		t.Errorf("PricePerShare = %.2f, want 1500", got.PricePerShare)
	}
	// BuyCost = 1000 * 1000 * (1 + 0.15/100) = 1_000_000 * 1.0015 = 1_001_500
	// CapitalGainPct = ((1_496_250 - 1_001_500) / 1_001_500) * 100
	expectedBuyCost := 1000.0 * 1000 * (1 + 0.15/100)
	expectedGain := ((1496250 - expectedBuyCost) / expectedBuyCost) * 100
	if math.Abs(got.CapitalGainPct-expectedGain) > 0.01 {
		t.Errorf("CapitalGainPct = %.4f, want ~%.4f", got.CapitalGainPct, expectedGain)
	}
	if got.NewPositionLots != 0 {
		t.Errorf("NewPositionLots = %d, want 0", got.NewPositionLots)
	}
	if got.NewPositionPct != 0 {
		t.Errorf("NewPositionPct = %.2f, want 0", got.NewPositionPct)
	}
	if got.NewAvgBuyPrice != 0 {
		t.Errorf("NewAvgBuyPrice = %.2f, want 0", got.NewAvgBuyPrice)
	}
}

func TestComputeSuggestionSellStop(t *testing.T) {
	input := EvaluateInput{
		Action:    ActionSellStop,
		StockData: &stock.Data{Ticker: "BBCA", Price: 800},
		Holding:   &portfolio.Holding{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 10},
		Portfolio: &portfolio.Portfolio{Capital: 10000000},
		AllHoldings: []*portfolio.Holding{
			{Ticker: "BBCA", AvgBuyPrice: 1000, Lots: 10},
		},
		BuyFeePct:  0.15,
		SellFeePct: 0.15,
		SellTaxPct: 0.10,
	}

	got, err := ComputeSuggestion(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Action != ActionSellStop {
		t.Errorf("Action = %q, want %q", got.Action, ActionSellStop)
	}
	// shares = 10 * 100 = 1000
	// GrossCost = 800 * 1000 = 800_000
	if got.GrossCost != 800000 {
		t.Errorf("GrossCost = %.2f, want 800000", got.GrossCost)
	}
	// Fee = 800_000 * 0.15 / 100 = 1200
	if got.Fee != 1200 {
		t.Errorf("Fee = %.2f, want 1200", got.Fee)
	}
	// Tax = 800_000 * 0.10 / 100 = 800
	if got.Tax != 800 {
		t.Errorf("Tax = %.2f, want 800", got.Tax)
	}
	// NetCost = 800_000 - 1200 - 800 = 798_000
	if got.NetCost != 798000 {
		t.Errorf("NetCost = %.2f, want 798000", got.NetCost)
	}
	// BuyCost = 1000 * 1000 * 1.0015 = 1_001_500
	// CapitalGainPct should be negative (loss)
	if got.CapitalGainPct >= 0 {
		t.Errorf("CapitalGainPct = %.4f, want negative (loss)", got.CapitalGainPct)
	}
	if got.NewPositionLots != 0 {
		t.Errorf("NewPositionLots = %d, want 0", got.NewPositionLots)
	}
}

func TestComputeSuggestionHold(t *testing.T) {
	input := EvaluateInput{
		Action:    ActionHold,
		StockData: &stock.Data{Ticker: "BBCA", Price: 1000},
		Portfolio: &portfolio.Portfolio{Capital: 10000000},
	}

	got, err := ComputeSuggestion(input)
	if got != nil {
		t.Errorf("expected nil suggestion, got %+v", got)
	}
	if !errors.Is(err, ErrHoldNoSuggestion) {
		t.Errorf("error = %v, want %v", err, ErrHoldNoSuggestion)
	}
}

func TestComputeSuggestionErrors(t *testing.T) {
	tests := []struct {
		name            string
		input           EvaluateInput
		wantErr         error
		wantErrContains string
	}{
		{
			name: "nil holding for AverageDown",
			input: EvaluateInput{
				Action:     ActionAverageDown,
				StockData:  &stock.Data{Ticker: "BBCA", Price: 800},
				Portfolio:  &portfolio.Portfolio{Capital: 10000000},
				Thresholds: Thresholds{MaxPositionPct: 20},
			},
			wantErr: ErrNoHolding,
		},
		{
			name: "nil holding for SellExit",
			input: EvaluateInput{
				Action:    ActionSellExit,
				StockData: &stock.Data{Ticker: "BBCA", Price: 1500},
				Portfolio: &portfolio.Portfolio{Capital: 10000000},
			},
			wantErr: ErrNoHolding,
		},
		{
			name: "nil holding for SellStop",
			input: EvaluateInput{
				Action:    ActionSellStop,
				StockData: &stock.Data{Ticker: "BBCA", Price: 800},
				Portfolio: &portfolio.Portfolio{Capital: 10000000},
			},
			wantErr: ErrNoHolding,
		},
		{
			name: "no room to buy position at max",
			input: EvaluateInput{
				Action:     ActionBuy,
				StockData:  &stock.Data{Ticker: "BBCA", Price: 1000},
				Holding:    &portfolio.Holding{Ticker: "BBCA", AvgBuyPrice: 900, Lots: 20},
				Portfolio:  &portfolio.Portfolio{Capital: 10000000},
				Thresholds: Thresholds{MaxPositionPct: 20},
				AllHoldings: []*portfolio.Holding{
					{Ticker: "BBCA", AvgBuyPrice: 900, Lots: 20},
				},
				BuyFeePct: 0.15,
			},
			wantErr:         ErrChecklistNotReady,
			wantErrContains: "no room to buy",
		},
		{
			name: "price too high for 1 lot",
			input: EvaluateInput{
				Action:      ActionBuy,
				StockData:   &stock.Data{Ticker: "BBCA", Price: 50000},
				Portfolio:   &portfolio.Portfolio{Capital: 1000000},
				Thresholds:  Thresholds{MaxPositionPct: 5},
				AllHoldings: []*portfolio.Holding{},
				BuyFeePct:   0.15,
			},
			wantErr:         ErrChecklistNotReady,
			wantErrContains: "insufficient budget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ComputeSuggestion(tt.input)
			if got != nil {
				t.Errorf("expected nil suggestion, got %+v", got)
			}
			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErrContains != "" {
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrContains)
				}
			}
		})
	}
}

func TestComputeSuggestionBuyWithOtherHoldings(t *testing.T) {
	input := EvaluateInput{
		Action:     ActionBuy,
		StockData:  &stock.Data{Ticker: "BBCA", Price: 1000},
		Portfolio:  &portfolio.Portfolio{Capital: 10000000},
		Thresholds: Thresholds{MaxPositionPct: 20},
		AllHoldings: []*portfolio.Holding{
			{Ticker: "BBRI", AvgBuyPrice: 500, Lots: 20},
		},
		BuyFeePct:  0.15,
		SellFeePct: 0.15,
		SellTaxPct: 0.10,
	}

	got, err := ComputeSuggestion(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Lots != 20 {
		t.Errorf("Lots = %d, want 20", got.Lots)
	}

	// totalPortfolioValue = BBRI (500 * 20 * 100 = 1_000_000) + new BBCA (1000 * 20 * 100 = 2_000_000) = 3_000_000
	// NewPositionPct = (2_000_000 / 3_000_000) * 100 = 66.67
	expectedPct := (2000000.0 / 3000000.0) * 100
	if math.Abs(got.NewPositionPct-expectedPct) > 0.01 {
		t.Errorf("NewPositionPct = %.4f, want ~%.4f", got.NewPositionPct, expectedPct)
	}
}
