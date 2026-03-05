package scraper

import "testing"

func TestFormatIDX(t *testing.T) {
	tests := []struct {
		name   string
		ticker string
		want   string
	}{
		{name: "plain ticker", ticker: "BBCA", want: "BBCA.JK"},
		{name: "already has suffix", ticker: "BBCA.JK", want: "BBCA.JK"},
		{name: "lowercase ticker", ticker: "bbca", want: "bbca.JK"},
		{name: "empty string", ticker: "", want: ".JK"},
		{name: "partial suffix", ticker: "BBCA.J", want: "BBCA.J.JK"},
		{name: "index ticker ^JKSE", ticker: "^JKSE", want: "^JKSE"},
		{name: "index ticker ^DJI", ticker: "^DJI", want: "^DJI"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatIDX(tt.ticker)
			if got != tt.want {
				t.Errorf("FormatIDX(%q) = %q, want %q", tt.ticker, got, tt.want)
			}
		})
	}
}
