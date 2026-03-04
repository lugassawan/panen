package usecase

import (
	"context"
	"errors"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"0.1.0", "0.2.0", -1},
		{"0.2.0", "0.1.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"0.9.9", "1.0.0", -1},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"dev", "0.1.0", -1},
		{"0.1.0", "dev", 1},
		{"dev", "dev", 0},
		{"0.0.1", "0.0.1", 0},
		{"1.0.0", "1.0.1", -1},
	}

	for _, tc := range tests {
		t.Run(tc.a+"_vs_"+tc.b, func(t *testing.T) {
			got := compareVersions(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

type mockReleaseChecker struct {
	info *ReleaseInfo
	err  error
}

func (m *mockReleaseChecker) LatestRelease(_ context.Context) (*ReleaseInfo, error) {
	return m.info, m.err
}

func TestUpdateServiceCheck(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		latest    *ReleaseInfo
		checkErr  error
		wantAvail bool
		wantErr   bool
	}{
		{
			name:    "update available",
			current: "0.1.0",
			latest: &ReleaseInfo{
				Version:    "0.2.0",
				ReleaseURL: "https://github.com/lugassawan/panen/releases/tag/v0.2.0",
			},
			wantAvail: true,
		},
		{
			name:    "up to date",
			current: "0.2.0",
			latest: &ReleaseInfo{
				Version:    "0.2.0",
				ReleaseURL: "https://github.com/lugassawan/panen/releases/tag/v0.2.0",
			},
			wantAvail: false,
		},
		{
			name:    "dev version always outdated",
			current: "dev",
			latest: &ReleaseInfo{
				Version:    "0.1.0",
				ReleaseURL: "https://github.com/lugassawan/panen/releases/tag/v0.1.0",
			},
			wantAvail: true,
		},
		{
			name:     "checker error",
			current:  "0.1.0",
			checkErr: errors.New("network error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			checker := &mockReleaseChecker{info: tc.latest, err: tc.checkErr}
			svc := NewUpdateService(checker, tc.current)

			result, err := svc.Check(context.Background())
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Available != tc.wantAvail {
				t.Errorf("Available = %v, want %v", result.Available, tc.wantAvail)
			}
			if result.CurrentVer != tc.current {
				t.Errorf("CurrentVer = %q, want %q", result.CurrentVer, tc.current)
			}
		})
	}
}

func TestUpdateServiceCurrentVersion(t *testing.T) {
	svc := NewUpdateService(nil, "1.2.3")
	if got := svc.CurrentVersion(); got != "1.2.3" {
		t.Errorf("CurrentVersion() = %q, want %q", got, "1.2.3")
	}
}
