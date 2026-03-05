package trailingstop

import (
	"testing"

	"github.com/lugassawan/panen/backend/domain/portfolio"
)

func TestStopPercentage(t *testing.T) {
	tests := []struct {
		name    string
		risk    portfolio.RiskProfile
		want    float64
		wantErr bool
	}{
		{name: "conservative", risk: portfolio.RiskProfileConservative, want: 9.0},
		{name: "moderate", risk: portfolio.RiskProfileModerate, want: 13.5},
		{name: "aggressive", risk: portfolio.RiskProfileAggressive, want: 21.5},
		{name: "invalid", risk: "INVALID", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StopPercentage(tt.risk)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("StopPercentage(%s) = %v, want %v", tt.risk, got, tt.want)
			}
		})
	}
}

func TestStopPrice(t *testing.T) {
	tests := []struct {
		name      string
		peakPrice float64
		stopPct   float64
		want      float64
	}{
		{name: "9% of 10000", peakPrice: 10000, stopPct: 9.0, want: 9100},
		{name: "13.5% of 10000", peakPrice: 10000, stopPct: 13.5, want: 8650},
		{name: "21.5% of 10000", peakPrice: 10000, stopPct: 21.5, want: 7850},
		{name: "zero peak", peakPrice: 0, stopPct: 10, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StopPrice(tt.peakPrice, tt.stopPct)
			if got != tt.want {
				t.Errorf("StopPrice(%v, %v) = %v, want %v", tt.peakPrice, tt.stopPct, got, tt.want)
			}
		})
	}
}

func TestIsTriggered(t *testing.T) {
	tests := []struct {
		name         string
		currentPrice float64
		stopPrice    float64
		want         bool
	}{
		{name: "below stop", currentPrice: 8000, stopPrice: 9100, want: true},
		{name: "at stop", currentPrice: 9100, stopPrice: 9100, want: true},
		{name: "above stop", currentPrice: 9500, stopPrice: 9100, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTriggered(tt.currentPrice, tt.stopPrice)
			if got != tt.want {
				t.Errorf("IsTriggered(%v, %v) = %v, want %v", tt.currentPrice, tt.stopPrice, got, tt.want)
			}
		})
	}
}

func TestUpdatePeak(t *testing.T) {
	tests := []struct {
		name         string
		currentPeak  float64
		currentPrice float64
		want         float64
	}{
		{name: "price higher than peak", currentPeak: 10000, currentPrice: 11000, want: 11000},
		{name: "price equal to peak", currentPeak: 10000, currentPrice: 10000, want: 10000},
		{name: "price lower than peak", currentPeak: 10000, currentPrice: 9000, want: 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdatePeak(tt.currentPeak, tt.currentPrice)
			if got != tt.want {
				t.Errorf("UpdatePeak(%v, %v) = %v, want %v", tt.currentPeak, tt.currentPrice, got, tt.want)
			}
		})
	}
}

func TestEvaluateFundamentals(t *testing.T) {
	tests := []struct {
		name          string
		roe           float64
		der           float64
		eps           float64
		wantTriggered []bool
	}{
		{
			name: "all healthy",
			roe:  15, der: 0.5, eps: 500,
			wantTriggered: []bool{false, false, false},
		},
		{
			name: "all deteriorated",
			roe:  5, der: 2.0, eps: -10,
			wantTriggered: []bool{true, true, true},
		},
		{
			name: "only roe low",
			roe:  8, der: 1.0, eps: 100,
			wantTriggered: []bool{true, false, false},
		},
		{
			name: "only der high",
			roe:  12, der: 1.6, eps: 100,
			wantTriggered: []bool{false, true, false},
		},
		{
			name: "eps zero",
			roe:  12, der: 0.5, eps: 0,
			wantTriggered: []bool{false, false, true},
		},
		{
			name: "boundary roe exactly 10",
			roe:  10, der: 1.0, eps: 100,
			wantTriggered: []bool{false, false, false},
		},
		{
			name: "boundary der exactly 1.5",
			roe:  15, der: 1.5, eps: 100,
			wantTriggered: []bool{false, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exits := EvaluateFundamentals(tt.roe, tt.der, tt.eps)
			if len(exits) != 3 {
				t.Fatalf("len(exits) = %d, want 3", len(exits))
			}
			for i, exit := range exits {
				if exit.Triggered != tt.wantTriggered[i] {
					t.Errorf(
						"exits[%d].Triggered = %v, want %v (key=%s)",
						i,
						exit.Triggered,
						tt.wantTriggered[i],
						exit.Key,
					)
				}
			}
		})
	}
}
