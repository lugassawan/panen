package crashplaybook

import "testing"

func TestDetectMarketCondition(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		peak     float64
		previous MarketCondition
		want     MarketCondition
	}{
		{name: "normal - no drawdown", price: 7500, peak: 7500, previous: MarketNormal, want: MarketNormal},
		{name: "normal - small dip", price: 7200, peak: 7500, previous: MarketNormal, want: MarketNormal},
		{name: "elevated - 5% drawdown", price: 7125, peak: 7500, previous: MarketNormal, want: MarketElevated},
		{name: "elevated - 8% drawdown", price: 6900, peak: 7500, previous: MarketNormal, want: MarketElevated},
		{name: "correction - 10% drawdown", price: 6750, peak: 7500, previous: MarketNormal, want: MarketCorrection},
		{name: "correction - 15% drawdown", price: 6375, peak: 7500, previous: MarketNormal, want: MarketCorrection},
		{name: "crash - 20% drawdown", price: 6000, peak: 7500, previous: MarketNormal, want: MarketCrash},
		{name: "crash - 30% drawdown", price: 5250, peak: 7500, previous: MarketNormal, want: MarketCrash},
		{name: "recovery from crash", price: 7000, peak: 7500, previous: MarketCrash, want: MarketRecovery},
		{name: "recovery from correction", price: 6900, peak: 7500, previous: MarketCorrection, want: MarketRecovery},
		{
			name:     "no recovery from normal - stays elevated",
			price:    7000,
			peak:     7500,
			previous: MarketNormal,
			want:     MarketElevated,
		},
		{
			name: "recovery to normal when past -5%", price: 7200, peak: 7500,
			previous: MarketRecovery, want: MarketNormal,
		},
		{
			name: "stays recovery between -5% and -10%", price: 7000, peak: 7500,
			previous: MarketRecovery, want: MarketRecovery,
		},
		{name: "still crash despite previous crash", price: 5500, peak: 7500, previous: MarketCrash, want: MarketCrash},
		{name: "zero peak returns normal", price: 100, peak: 0, previous: MarketNormal, want: MarketNormal},
		{name: "exact threshold -5%", price: 7125, peak: 7500, previous: MarketNormal, want: MarketElevated},
		{name: "exact threshold -10%", price: 6750, peak: 7500, previous: MarketNormal, want: MarketCorrection},
		{name: "exact threshold -20%", price: 6000, peak: 7500, previous: MarketNormal, want: MarketCrash},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectMarketCondition(tt.price, tt.peak, tt.previous)
			if got != tt.want {
				t.Errorf("DetectMarketCondition(%v, %v, %v) = %v, want %v",
					tt.price, tt.peak, tt.previous, got, tt.want)
			}
		})
	}
}

func TestDrawdownPct(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		peak  float64
		want  float64
	}{
		{name: "no drawdown", price: 7500, peak: 7500, want: 0},
		{name: "10% drawdown", price: 6750, peak: 7500, want: -10},
		{name: "zero peak", price: 100, peak: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DrawdownPct(tt.price, tt.peak)
			if got != tt.want {
				t.Errorf("DrawdownPct(%v, %v) = %v, want %v", tt.price, tt.peak, got, tt.want)
			}
		})
	}
}
