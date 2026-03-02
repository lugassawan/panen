package portfolio

import (
	"errors"
	"testing"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		input   string
		want    Mode
		wantErr error
	}{
		{input: "VALUE", want: ModeValue},
		{input: "DIVIDEND", want: ModeDividend},
		{input: "INVALID", wantErr: ErrInvalidMode},
		{input: "", wantErr: ErrInvalidMode},
		{input: "value", wantErr: ErrInvalidMode},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseMode(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMode(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseMode(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseRiskProfile(t *testing.T) {
	tests := []struct {
		input   string
		want    RiskProfile
		wantErr error
	}{
		{input: "CONSERVATIVE", want: RiskProfileConservative},
		{input: "MODERATE", want: RiskProfileModerate},
		{input: "AGGRESSIVE", want: RiskProfileAggressive},
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
