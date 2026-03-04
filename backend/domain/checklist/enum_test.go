package checklist

import (
	"errors"
	"testing"
)

func TestParseActionType(t *testing.T) {
	tests := []struct {
		input   string
		want    ActionType
		wantErr error
	}{
		{input: "BUY", want: ActionBuy},
		{input: "AVERAGE_DOWN", want: ActionAverageDown},
		{input: "AVERAGE_UP", want: ActionAverageUp},
		{input: "SELL_EXIT", want: ActionSellExit},
		{input: "SELL_STOP", want: ActionSellStop},
		{input: "HOLD", want: ActionHold},
		{input: "INVALID", wantErr: ErrInvalidAction},
		{input: "", wantErr: ErrInvalidAction},
		{input: "buy", wantErr: ErrInvalidAction},
		{input: "Buy", wantErr: ErrInvalidAction},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseActionType(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseActionType(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseActionType(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseActionType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
