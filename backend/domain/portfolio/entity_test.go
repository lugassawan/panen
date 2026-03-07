package portfolio

import "testing"

func TestHoldingComputeAvgBuyPrice(t *testing.T) {
	tests := []struct {
		name     string
		holding  Holding
		newPrice float64
		newLots  int
		want     float64
	}{
		{
			name:     "equal lots same price",
			holding:  Holding{AvgBuyPrice: 1000, Lots: 10},
			newPrice: 1000,
			newLots:  10,
			want:     1000,
		},
		{
			name:     "average down",
			holding:  Holding{AvgBuyPrice: 1000, Lots: 10},
			newPrice: 800,
			newLots:  10,
			want:     900,
		},
		{
			name:     "average up",
			holding:  Holding{AvgBuyPrice: 1000, Lots: 10},
			newPrice: 1200,
			newLots:  10,
			want:     1100,
		},
		{
			name:     "unequal lots",
			holding:  Holding{AvgBuyPrice: 1000, Lots: 10},
			newPrice: 500,
			newLots:  20,
			want:     (1000*10 + 500*20) / 30.0,
		},
		{
			name:     "single new lot",
			holding:  Holding{AvgBuyPrice: 1000, Lots: 99},
			newPrice: 2000,
			newLots:  1,
			want:     (1000*99 + 2000*1) / 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.holding.ComputeAvgBuyPrice(tt.newPrice, tt.newLots)
			if got != tt.want {
				t.Errorf("ComputeAvgBuyPrice(%v, %d) = %v, want %v", tt.newPrice, tt.newLots, got, tt.want)
			}
		})
	}
}

// feeCalcCase holds inputs and expected output for transaction cost functions
// (ComputeBuyFee, ComputeSellFee, ComputeSellTax) which all share the same
// formula: price * lots * 100 * pct / 100.
type feeCalcCase struct {
	name  string
	price float64
	lots  int
	pct   float64
	want  float64
}

func runFeeCalcTests(t *testing.T, fn func(float64, int, float64) float64, cases []feeCalcCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(tt.price, tt.lots, tt.pct)
			if got != tt.want {
				t.Errorf("(%v, %d, %v) = %v, want %v", tt.price, tt.lots, tt.pct, got, tt.want)
			}
		})
	}
}

func TestComputeBuyFee(t *testing.T) {
	runFeeCalcTests(t, ComputeBuyFee, []feeCalcCase{
		{"standard fee", 1000, 10, 0.15, 1000 * 1000 * 0.15 / 100},
		{"zero fee", 1000, 10, 0, 0},
		{"single lot", 500, 1, 0.15, 500 * 100 * 0.15 / 100},
		{"high fee percentage", 2000, 5, 0.25, 2000 * 500 * 0.25 / 100},
	})
}

func TestComputeSellFee(t *testing.T) {
	runFeeCalcTests(t, ComputeSellFee, []feeCalcCase{
		{"standard fee", 1000, 10, 0.25, 1000 * 1000 * 0.25 / 100},
		{"zero fee", 1000, 10, 0, 0},
		{"single lot", 500, 1, 0.25, 500 * 100 * 0.25 / 100},
		{"high fee percentage", 2000, 5, 0.35, 2000 * 500 * 0.35 / 100},
	})
}

func TestComputeSellTax(t *testing.T) {
	runFeeCalcTests(t, ComputeSellTax, []feeCalcCase{
		{"standard tax", 1000, 10, 0.10, 1000 * 1000 * 0.10 / 100},
		{"zero tax", 1000, 10, 0, 0},
		{"single lot", 500, 1, 0.10, 500 * 100 * 0.10 / 100},
		{"high tax percentage", 2000, 5, 0.15, 2000 * 500 * 0.15 / 100},
	})
}
