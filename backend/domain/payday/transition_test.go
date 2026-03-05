package payday

import (
	"fmt"
	"testing"
)

func TestValidTransition(t *testing.T) {
	tests := []struct {
		from Status
		to   Status
		want bool
	}{
		// Valid transitions
		{from: StatusScheduled, to: StatusPending, want: true},
		{from: StatusPending, to: StatusConfirmed, want: true},
		{from: StatusPending, to: StatusDeferred, want: true},
		{from: StatusPending, to: StatusSkipped, want: true},
		{from: StatusDeferred, to: StatusPending, want: true},

		// Invalid transitions
		{from: StatusScheduled, to: StatusConfirmed, want: false},
		{from: StatusScheduled, to: StatusDeferred, want: false},
		{from: StatusScheduled, to: StatusSkipped, want: false},
		{from: StatusPending, to: StatusScheduled, want: false},
		{from: StatusConfirmed, to: StatusPending, want: false},
		{from: StatusConfirmed, to: StatusScheduled, want: false},
		{from: StatusSkipped, to: StatusPending, want: false},
		{from: StatusSkipped, to: StatusScheduled, want: false},
		{from: StatusDeferred, to: StatusConfirmed, want: false},
		{from: StatusDeferred, to: StatusSkipped, want: false},

		// Self-transitions
		{from: StatusScheduled, to: StatusScheduled, want: false},
		{from: StatusPending, to: StatusPending, want: false},
		{from: StatusConfirmed, to: StatusConfirmed, want: false},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s->%s", tt.from, tt.to)
		t.Run(name, func(t *testing.T) {
			got := ValidTransition(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("ValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
