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

func TestComputeBuyFee(t *testing.T) {
	tests := []struct {
		name      string
		price     float64
		lots      int
		buyFeePct float64
		want      float64
	}{
		{
			name:      "standard fee",
			price:     1000,
			lots:      10,
			buyFeePct: 0.15,
			want:      1000 * 1000 * 0.15 / 100, // 1000 shares * 1000 price * 0.15%
		},
		{
			name:      "zero fee",
			price:     1000,
			lots:      10,
			buyFeePct: 0,
			want:      0,
		},
		{
			name:      "single lot",
			price:     500,
			lots:      1,
			buyFeePct: 0.15,
			want:      500 * 100 * 0.15 / 100,
		},
		{
			name:      "high fee percentage",
			price:     2000,
			lots:      5,
			buyFeePct: 0.25,
			want:      2000 * 500 * 0.25 / 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeBuyFee(tt.price, tt.lots, tt.buyFeePct)
			if got != tt.want {
				t.Errorf("ComputeBuyFee(%v, %d, %v) = %v, want %v", tt.price, tt.lots, tt.buyFeePct, got, tt.want)
			}
		})
	}
}
