package dividend

import "testing"

func TestDetermineIndicator(t *testing.T) {
	tests := []struct {
		name  string
		input IndicatorInput
		want  Indicator
	}{
		{
			name: "overvalued by price above exit",
			input: IndicatorInput{
				HasHolding: true, Price: 5000, ExitTarget: 4500,
				DividendYield: 5, PayoutRatio: 50,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorOvervalued,
		},
		{
			name: "overvalued by high payout ratio",
			input: IndicatorInput{
				HasHolding: true, Price: 3000, ExitTarget: 5000,
				DividendYield: 5, PayoutRatio: 80,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorOvervalued,
		},
		{
			name: "buy zone no holding price below entry",
			input: IndicatorInput{
				HasHolding: false, Price: 2500, EntryPrice: 3000, ExitTarget: 5000,
				DividendYield: 5, PayoutRatio: 50,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorBuyZone,
		},
		{
			name: "no holding but price above entry",
			input: IndicatorInput{
				HasHolding: false, Price: 3500, EntryPrice: 3000, ExitTarget: 5000,
				DividendYield: 5, PayoutRatio: 50,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorHold,
		},
		{
			name: "no holding yield below min",
			input: IndicatorInput{
				HasHolding: false, Price: 2500, EntryPrice: 3000, ExitTarget: 5000,
				DividendYield: 2, PayoutRatio: 50,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorHold,
		},
		{
			name: "average up criteria met",
			input: IndicatorInput{
				HasHolding: true, Price: 3000, ExitTarget: 5000,
				DividendYield: 5, PayoutRatio: 50, PositionPct: 10,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorAverageUp,
		},
		{
			name: "hold when yield too low for average up",
			input: IndicatorInput{
				HasHolding: true, Price: 3000, ExitTarget: 5000,
				DividendYield: 2, PayoutRatio: 50, PositionPct: 10,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorHold,
		},
		{
			name: "hold when position at max",
			input: IndicatorInput{
				HasHolding: true, Price: 3000, ExitTarget: 5000,
				DividendYield: 5, PayoutRatio: 50, PositionPct: 20,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorHold,
		},
		{
			name: "overvalued takes priority over average up",
			input: IndicatorInput{
				HasHolding: true, Price: 6000, ExitTarget: 5000,
				DividendYield: 5, PayoutRatio: 50, PositionPct: 10,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorOvervalued,
		},
		{
			name: "boundary price equals entry for buy zone",
			input: IndicatorInput{
				HasHolding: false, Price: 3000, EntryPrice: 3000, ExitTarget: 5000,
				DividendYield: 3, PayoutRatio: 50,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorBuyZone,
		},
		{
			name: "boundary price equals exit is overvalued",
			input: IndicatorInput{
				HasHolding: true, Price: 5000, ExitTarget: 5000,
				DividendYield: 5, PayoutRatio: 50,
				MinDY: 3, MaxPayoutRatio: 75, MaxPositionPct: 20,
			},
			want: IndicatorOvervalued,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineIndicator(tt.input)
			if got != tt.want {
				t.Errorf("DetermineIndicator() = %q, want %q", got, tt.want)
			}
		})
	}
}
