package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLatestRelease(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantVer    string
		wantURL    string
		wantErrStr string
	}{
		{
			name:   "success",
			status: http.StatusOK,
			body: `{"tag_name":"v0.2.0",` +
				`"html_url":"https://github.com/lugassawan/panen/releases/tag/v0.2.0",` +
				`"name":"v0.2.0"}`,
			wantVer: "0.2.0",
			wantURL: "https://github.com/lugassawan/panen/releases/tag/v0.2.0",
		},
		{
			name:       "non-200 status",
			status:     http.StatusNotFound,
			body:       `{"message":"Not Found"}`,
			wantErrStr: "github API returned status 404",
		},
		{
			name:       "malformed JSON",
			status:     http.StatusOK,
			body:       `{invalid`,
			wantErrStr: "decode response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := NewClient(WithAPIURL(srv.URL))
			rel, err := client.LatestRelease(context.Background())

			if tc.wantErrStr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrStr)
				}
				if got := err.Error(); !strings.Contains(got, tc.wantErrStr) {
					t.Fatalf("error %q does not contain %q", got, tc.wantErrStr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := rel.Version(); got != tc.wantVer {
				t.Errorf("Version() = %q, want %q", got, tc.wantVer)
			}
			if got := rel.HTMLURL; got != tc.wantURL {
				t.Errorf("HTMLURL = %q, want %q", got, tc.wantURL)
			}
		})
	}
}

func TestReleaseVersion(t *testing.T) {
	tests := []struct {
		tag  string
		want string
	}{
		{"v1.2.3", "1.2.3"},
		{"0.1.0", "0.1.0"},
		{"v0.0.1", "0.0.1"},
	}
	for _, tc := range tests {
		r := &Release{TagName: tc.tag}
		if got := r.Version(); got != tc.want {
			t.Errorf("Release{TagName: %q}.Version() = %q, want %q", tc.tag, got, tc.want)
		}
	}
}
