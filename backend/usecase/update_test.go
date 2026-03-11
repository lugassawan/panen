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

func TestCheckReturnsReleaseNotes(t *testing.T) {
	checker := &mockReleaseChecker{
		info: &ReleaseInfo{
			Version:    "1.1.0",
			ReleaseURL: "https://github.com/lugassawan/panen/releases/tag/v1.1.0",
			ReleaseNotes: "## What's Changed\n" +
				"- feat: add cool feature (#10)\n" +
				"- fix: broken thing (#11)\n" +
				"- chore: bump deps (#12)",
		},
	}
	svc := NewUpdateService(checker, "1.0.0")
	result, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "- Add cool feature (#10)\n- Broken thing (#11)"
	if result.ReleaseNotes != want {
		t.Errorf("ReleaseNotes = %q, want %q", result.ReleaseNotes, want)
	}
}

func TestCleanReleaseNotes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "extracts feat and fix lines",
			input: "## What's Changed\n### Features\n" +
				"- feat: add version bump script (#135)\n" +
				"### Bug Fixes\n" +
				"- fix: capitalize macOS app bundle name (#131)\n" +
				"### Other Changes\n" +
				"- chore: bump version to 1.0.1 (#134)\n" +
				"### Checksums",
			want: "- Add version bump script (#135)\n- Capitalize macOS app bundle name (#131)",
		},
		{
			name:  "empty body",
			input: "",
			want:  "",
		},
		{
			name:  "no feat or fix lines",
			input: "## What's Changed\n- chore: bump deps\n- docs: update readme",
			want:  "",
		},
		{
			name:  "asterisk bullets",
			input: "* feat: something new\n* fix: something broken",
			want:  "- Something new\n- Something broken",
		},
		{
			name:  "already capitalized",
			input: "- feat: Add feature X",
			want:  "- Add feature X",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cleanReleaseNotes(tc.input)
			if got != tc.want {
				t.Errorf("cleanReleaseNotes() = %q, want %q", got, tc.want)
			}
		})
	}
}
