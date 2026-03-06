package dividend

import (
	"math"
	"testing"
	"time"
)

func TestDeriveAnnualDPS(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		dyPct float64
		want  float64
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
				{PositionValue: 10_000_000, AnnualDPS: 200, Lots: 10}, // income = 200*10*100 = 200000
				{PositionValue: 20_000_000, AnnualDPS: 150, Lots: 20}, // income = 150*20*100 = 300000
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

func TestYoCProgression(t *testing.T) {
	events := []DividendEvent{
		{ExDate: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50},
		{ExDate: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC), Amount: 60},
		{ExDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 55},
	}
	avgBuyPrice := 2000.0

	points := YoCProgression(events, avgBuyPrice)
	if len(points) != 3 {
		t.Fatalf("got %d points, want 3", len(points))
	}

	// First point: trailing 12m DPS = 50, YoC = 50/2000*100 = 2.5
	if !almostEqual(points[0].YoC, 2.5) {
		t.Errorf("points[0].YoC = %v, want 2.5", points[0].YoC)
	}

	// Second point: trailing 12m DPS = 50+60 = 110, YoC = 110/2000*100 = 5.5
	if !almostEqual(points[1].YoC, 5.5) {
		t.Errorf("points[1].YoC = %v, want 5.5", points[1].YoC)
	}

	// Third point: trailing 12m (2023-03-16 to 2024-03-15) DPS = 60+55 = 115, YoC = 115/2000*100 = 5.75
	if !almostEqual(points[2].YoC, 5.75) {
		t.Errorf("points[2].YoC = %v, want 5.75", points[2].YoC)
	}
}

func TestYoCProgressionZeroAvgBuyPrice(t *testing.T) {
	events := []DividendEvent{
		{ExDate: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), Amount: 50},
	}
	points := YoCProgression(events, 0)
	if points != nil {
		t.Errorf("got %v, want nil", points)
	}
}

func TestYoCProgressionEmpty(t *testing.T) {
	points := YoCProgression(nil, 2000)
	if points != nil {
		t.Errorf("got %v, want nil", points)
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}
