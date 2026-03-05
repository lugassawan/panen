package crashplaybook

import "testing"

func TestComputeResponseLevels(t *testing.T) {
	entryPrice := 1000.0
	low52Week := 600.0
	deployPcts := [3]float64{30, 40, 30}

	levels := ComputeResponseLevels(entryPrice, low52Week, deployPcts)
	if len(levels) != 3 {
		t.Fatalf("expected 3 levels, got %d", len(levels))
	}

	tests := []struct {
		name         string
		level        CrashLevel
		triggerPrice float64
		deployPct    float64
	}{
		{name: "normal dip", level: LevelNormalDip, triggerPrice: 1000, deployPct: 30},
		{name: "crash", level: LevelCrash, triggerPrice: 800, deployPct: 40},
		{name: "extreme", level: LevelExtreme, triggerPrice: 630, deployPct: 30},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if levels[i].Level != tt.level {
				t.Errorf("level %d: got level %v, want %v", i, levels[i].Level, tt.level)
			}
			if levels[i].TriggerPrice != tt.triggerPrice {
				t.Errorf("level %d: got trigger %v, want %v", i, levels[i].TriggerPrice, tt.triggerPrice)
			}
			if levels[i].DeployPct != tt.deployPct {
				t.Errorf("level %d: got deploy %v, want %v", i, levels[i].DeployPct, tt.deployPct)
			}
		})
	}
}

func TestDetermineActiveLevel(t *testing.T) {
	levels := []ResponseLevel{
		{Level: LevelNormalDip, TriggerPrice: 1000},
		{Level: LevelCrash, TriggerPrice: 800},
		{Level: LevelExtreme, TriggerPrice: 630},
	}

	normalDip := LevelNormalDip
	crash := LevelCrash
	extreme := LevelExtreme

	tests := []struct {
		name  string
		price float64
		want  *CrashLevel
	}{
		{name: "above all levels", price: 1100, want: nil},
		{name: "at normal dip", price: 1000, want: &normalDip},
		{name: "between normal dip and crash", price: 900, want: &normalDip},
		{name: "at crash", price: 800, want: &crash},
		{name: "between crash and extreme", price: 700, want: &crash},
		{name: "at extreme", price: 630, want: &extreme},
		{name: "below extreme", price: 500, want: &extreme},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineActiveLevel(tt.price, levels)
			if tt.want == nil && got != nil {
				t.Errorf("DetermineActiveLevel(%v) = %v, want nil", tt.price, *got)
			}
			if tt.want != nil && (got == nil || *got != *tt.want) {
				t.Errorf("DetermineActiveLevel(%v) = %v, want %v", tt.price, got, *tt.want)
			}
		})
	}
}
