package valuation

import (
	"errors"
	"testing"
)

func TestParseRiskProfile(t *testing.T) {
	tests := []struct {
		input   string
		want    RiskProfile
		wantErr error
	}{
		{input: "CONSERVATIVE", want: RiskConservative},
		{input: "MODERATE", want: RiskModerate},
		{input: "AGGRESSIVE", want: RiskAggressive},
		{input: "INVALID", wantErr: ErrInvalidRisk},
		{input: "", wantErr: ErrInvalidRisk},
		{input: "conservative", wantErr: ErrInvalidRisk},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRiskProfile(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseRiskProfile(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRiskProfile(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseRiskProfile(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
