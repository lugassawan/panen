package dividend

import (
	"math"
	"testing"
)

func TestDeriveAnnualDPS(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		dyPct    float64
		want     float64
	}{
		{name: "normal", price: 4000, dyPct: 5, want: 200},
		{name: "high yield", price: 2000, dyPct: 10, want: 200},
		{name: "zero price", price: 0, dyPct: 5, want: 0},
		{name: "negative price", price: -100, dyPct: 5, want: 0},
		{name: "zero yield", price: 4000, dyPct: 0, want: 0},
		{name: "negative yield", price: 4000, dyPct: -1, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveAnnualDPS(tt.price, tt.dyPct)
			if got != tt.want {
				t.Errorf("DeriveAnnualDPS(%v, %v) = %v, want %v", tt.price, tt.dyPct, got, tt.want)
			}
		})
	}
}

func TestYieldOnCost(t *testing.T) {
	tests := []struct {
		name        string
		annualDPS   float64
		avgBuyPrice float64
		want        float64
	}{
		{name: "normal", annualDPS: 200, avgBuyPrice: 3000, want: 200.0 / 3000 * 100},
		{name: "high yoc", annualDPS: 300, avgBuyPrice: 2000, want: 15},
		{name: "zero avg price", annualDPS: 200, avgBuyPrice: 0, want: 0},
		{name: "negative avg price", annualDPS: 200, avgBuyPrice: -100, want: 0},
		{name: "zero dps", annualDPS: 0, avgBuyPrice: 3000, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := YieldOnCost(tt.annualDPS, tt.avgBuyPrice)
			if !almostEqual(got, tt.want) {
				t.Errorf("YieldOnCost(%v, %v) = %v, want %v", tt.annualDPS, tt.avgBuyPrice, got, tt.want)
			}
		})
	}
}

func TestProjectedYoC(t *testing.T) {
	tests := []struct {
		name        string
		annualDPS   float64
		avgBuyPrice float64
		currentLots int
		newPrice    float64
		newLots     int
		want        float64
	}{
		{
			name:        "average up at higher price",
			annualDPS:   200,
			avgBuyPrice: 3000,
			currentLots: 10,
			newPrice:    4000,
			newLots:     5,
			// new avg = (10*100*3000 + 5*100*4000) / (15*100) = 5000000/1500 = 3333.33...
			// projected yoc = 200 / 3333.33 * 100 = 6.0
			want: 6.0,
		},
		{
			name:        "average up at same price",
			annualDPS:   200,
			avgBuyPrice: 3000,
			currentLots: 10,
			newPrice:    3000,
			newLots:     10,
			want:        200.0 / 3000 * 100,
		},
		{
			name:        "zero current lots",
			annualDPS:   200,
			avgBuyPrice: 3000,
			currentLots: 0,
			newPrice:    4000,
			newLots:     5,
			// new avg = (0 + 5*100*4000) / (5*100) = 4000
			want: 200.0 / 4000 * 100,
		},
		{
			name:        "invalid new lots",
			annualDPS:   200,
			avgBuyPrice: 3000,
			currentLots: 10,
			newPrice:    4000,
			newLots:     0,
			want:        0,
		},
		{
			name:        "negative current lots",
			annualDPS:   200,
			avgBuyPrice: 3000,
			currentLots: -1,
			newPrice:    4000,
			newLots:     5,
			want:        0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProjectedYoC(tt.annualDPS, tt.avgBuyPrice, tt.currentLots, tt.newPrice, tt.newLots)
			if !almostEqual(got, tt.want) {
				t.Errorf("ProjectedYoC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPortfolioYield(t *testing.T) {
	tests := []struct {
		name  string
		items []PortfolioYieldItem
		want  float64
	}{
		{
			name: "two holdings",
			items: []PortfolioYieldItem{
				{PositionValue: 10_000_000, AnnualDPS: 200, Lots: 10},  // income = 200*10*100 = 200000
				{PositionValue: 20_000_000, AnnualDPS: 150, Lots: 20},  // income = 150*20*100 = 300000
			},
			// total value = 30M, total income = 500000, yield = 500000/30M*100 = 1.6667
			want: 500000.0 / 30_000_000 * 100,
		},
		{
			name:  "empty",
			items: nil,
			want:  0,
		},
		{
			name: "zero position value skipped",
			items: []PortfolioYieldItem{
				{PositionValue: 0, AnnualDPS: 200, Lots: 10},
				{PositionValue: 10_000_000, AnnualDPS: 100, Lots: 5},
			},
			want: 100.0 * 5 * 100 / 10_000_000 * 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PortfolioYield(tt.items)
			if !almostEqual(got, tt.want) {
				t.Errorf("PortfolioYield() = %v, want %v", got, tt.want)
			}
		})
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}
