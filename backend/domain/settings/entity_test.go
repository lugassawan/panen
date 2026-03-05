package settings

import "testing"

func TestDefaultRefreshSettings(t *testing.T) {
	s := DefaultRefreshSettings()

	if !s.AutoRefreshEnabled {
		t.Error("expected AutoRefreshEnabled to be true")
	}
	if s.IntervalMinutes != 720 {
		t.Errorf("IntervalMinutes = %d, want %d", s.IntervalMinutes, 720)
	}
	if !s.LastRefreshedAt.IsZero() {
		t.Error("expected LastRefreshedAt to be zero for defaults")
	}
}
