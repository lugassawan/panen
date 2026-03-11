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
		wantBody   string
		wantErrStr string
	}{
		{
			name:   "success",
			status: http.StatusOK,
			body: `{"tag_name":"v0.2.0",` +
				`"html_url":"https://github.com/lugassawan/panen/releases/tag/v0.2.0",` +
				`"name":"v0.2.0",` +
				`"body":"## What's Changed\n- feat: cool feature",` +
				`"assets":[{"name":"test.zip",` +
				`"browser_download_url":"https://example.com/test.zip",` +
				`"size":12345}]}`,
			wantVer:  "0.2.0",
			wantURL:  "https://github.com/lugassawan/panen/releases/tag/v0.2.0",
			wantBody: "## What's Changed\n- feat: cool feature",
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
			if tc.wantBody != "" {
				if got := rel.Body; got != tc.wantBody {
					t.Errorf("Body = %q, want %q", got, tc.wantBody)
				}
			}
		})
	}
}

func TestLatestReleaseAssets(t *testing.T) {
	darwinURL := "https://github.com/lugassawan/panen/" +
		"releases/download/v0.3.0/panen-darwin-universal.zip"
	checksumURL := "https://github.com/lugassawan/panen/" +
		"releases/download/v0.3.0/SHA256SUMS.txt"
	body := `{
		"tag_name":"v0.3.0",
		"html_url":"https://github.com/lugassawan/panen/releases/tag/v0.3.0",
		"name":"v0.3.0",
		"assets":[
			{"name":"panen-darwin-universal.zip",
			 "browser_download_url":"` + darwinURL + `",
			 "size":50000},
			{"name":"SHA256SUMS.txt",
			 "browser_download_url":"` + checksumURL + `",
			 "size":256}
		]
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := NewClient(WithAPIURL(srv.URL))
	rel, err := client.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rel.Assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(rel.Assets))
	}
	if rel.Assets[0].Name != "panen-darwin-universal.zip" {
		t.Errorf("asset name = %q, want %q", rel.Assets[0].Name, "panen-darwin-universal.zip")
	}
	if rel.Assets[0].Size != 50000 {
		t.Errorf("asset size = %d, want %d", rel.Assets[0].Size, 50000)
	}
	if rel.Assets[1].Name != "SHA256SUMS.txt" {
		t.Errorf("asset name = %q, want %q", rel.Assets[1].Name, "SHA256SUMS.txt")
	}
}

func TestDownloadAssetBlocksNonReleaseURL(t *testing.T) {
	client := NewClient()
	var buf strings.Builder
	err := client.DownloadAsset(context.Background(), "https://evil.com/malware.zip", &buf, nil)
	if err == nil {
		t.Fatal("expected error for non-release URL")
	}
	if !strings.Contains(err.Error(), "blocked non-release URL") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCountingWriter(t *testing.T) {
	var buf strings.Builder
	var calls []int64
	cw := &countingWriter{
		dest:  &buf,
		total: 10,
		progressFn: func(downloaded, total int64) {
			calls = append(calls, downloaded)
		},
	}
	_, _ = cw.Write([]byte("hello"))
	_, _ = cw.Write([]byte("world"))
	if buf.String() != "helloworld" {
		t.Errorf("written = %q, want %q", buf.String(), "helloworld")
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 progress calls, got %d", len(calls))
	}
	if calls[0] != 5 || calls[1] != 10 {
		t.Errorf("progress calls = %v, want [5 10]", calls)
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
