package payday

import (
	"errors"
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input   string
		want    Status
		wantErr error
	}{
		{input: "SCHEDULED", want: StatusScheduled},
		{input: "PENDING", want: StatusPending},
		{input: "CONFIRMED", want: StatusConfirmed},
		{input: "DEFERRED", want: StatusDeferred},
		{input: "SKIPPED", want: StatusSkipped},
		{input: "INVALID", wantErr: ErrInvalidStatus},
		{input: "", wantErr: ErrInvalidStatus},
		{input: "scheduled", wantErr: ErrInvalidStatus},
		{input: "Pending", wantErr: ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseStatus(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseStatus(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseStatus(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFlowType(t *testing.T) {
	tests := []struct {
		input   string
		want    FlowType
		wantErr error
	}{
		{input: "INITIAL", want: FlowTypeInitial},
		{input: "MONTHLY", want: FlowTypeMonthly},
		{input: "DIVIDEND", want: FlowTypeDividend},
		{input: "SALE", want: FlowTypeSale},
		{input: "INVALID", wantErr: ErrInvalidFlowType},
		{input: "", wantErr: ErrInvalidFlowType},
		{input: "initial", wantErr: ErrInvalidFlowType},
		{input: "Monthly", wantErr: ErrInvalidFlowType},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFlowType(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseFlowType(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFlowType(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseFlowType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
