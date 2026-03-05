package dividend

import (
	"math"
	"testing"
)

func TestScore(t *testing.T) {
	tests := []struct {
		name  string
		input ScoreInput
		want  float64
	}{
		{
			name: "perfect candidate",
			input: ScoreInput{
				DY: 10, MinDY: 3,
				PayoutRatio: 30, MaxPayoutRatio: 75,
				Price: 2000, EntryPrice: 4000,
				PositionPct: 0, MaxPositionPct: 20,
			},
			// yield premium: (10-3)/3 * 40 = 93.3 → capped at 40
			// payout margin: (75-30)/75 * 20 = 12.0
			// valuation: (4000-2000)/4000 * 20 = 10.0
			// headroom: (20-0)/20 * 20 = 20.0
			want: 40 + 12 + 10 + 20,
		},
		{
			name: "yield below minimum",
			input: ScoreInput{
				DY: 2, MinDY: 3,
				PayoutRatio: 30, MaxPayoutRatio: 75,
				Price: 2000, EntryPrice: 4000,
				PositionPct: 0, MaxPositionPct: 20,
			},
			// yield premium: 0 (below min)
			want: 0 + 12 + 10 + 20,
		},
		{
			name: "price above entry no valuation bonus",
			input: ScoreInput{
				DY: 5, MinDY: 3,
				PayoutRatio: 50, MaxPayoutRatio: 75,
				Price: 5000, EntryPrice: 4000,
				PositionPct: 5, MaxPositionPct: 20,
			},
			// yield premium: (5-3)/3 * 40 = 26.667
			// payout margin: (75-50)/75 * 20 = 6.667
			// valuation: 0 (price > entry)
			// headroom: (20-5)/20 * 20 = 15.0
			want: 26.667 + 6.667 + 0 + 15,
		},
		{
			name: "zero minimums",
			input: ScoreInput{
				DY: 5, MinDY: 0,
				PayoutRatio: 50, MaxPayoutRatio: 0,
				Price: 3000, EntryPrice: 0,
				PositionPct: 10, MaxPositionPct: 0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.input)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("Score() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRank(t *testing.T) {
	items := []RankItem{
		{Ticker: "BBCA", Score: 30},
		{Ticker: "TLKM", Score: 50},
		{Ticker: "ASII", Score: 40},
	}
	ranked := Rank(items)
	if ranked[0].Ticker != "TLKM" || ranked[1].Ticker != "ASII" || ranked[2].Ticker != "BBCA" {
		t.Errorf("Rank() order = %v %v %v, want TLKM ASII BBCA",
			ranked[0].Ticker, ranked[1].Ticker, ranked[2].Ticker)
	}
}

func TestRankStableOrder(t *testing.T) {
	items := []RankItem{
		{Ticker: "AAAA", Score: 50},
		{Ticker: "BBBB", Score: 50},
	}
	ranked := Rank(items)
	if ranked[0].Ticker != "AAAA" || ranked[1].Ticker != "BBBB" {
		t.Errorf("Rank() should be stable, got %v %v", ranked[0].Ticker, ranked[1].Ticker)
	}
}
