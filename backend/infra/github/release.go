package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultAPIURL          = "https://api.github.com"
	defaultRepo            = "lugassawan/panen"
	maxResponseSize        = 1 << 20 // 1 MB
	downloadTimeout        = 5 * time.Minute
	allowedDownloadURLBase = "https://github.com/lugassawan/panen/releases/"
)

// Asset holds a single downloadable file attached to a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Release holds the relevant fields from a GitHub release response.
type Release struct {
	TagName string  `json:"tag_name"`
	HTMLURL string  `json:"html_url"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

// Version returns the tag name without the leading "v" prefix.
func (r *Release) Version() string {
	return strings.TrimPrefix(r.TagName, "v")
}

// Client fetches release information from the GitHub API.
type Client struct {
	http   *http.Client
	apiURL string
	repo   string
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.http = c }
}

// WithAPIURL overrides the GitHub API base URL (useful for tests).
func WithAPIURL(url string) Option {
	return func(cl *Client) { cl.apiURL = url }
}

// NewClient creates a new GitHub release client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		http:   &http.Client{Timeout: 10 * time.Second},
		apiURL: defaultAPIURL,
		repo:   defaultRepo,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// LatestRelease fetches the latest release from the configured repository.
func (c *Client) LatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", c.apiURL, c.repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req) //nolint:gosec // URL is constructed from controlled apiURL constant
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &rel, nil
}

// DownloadAsset downloads a release asset from the given URL to dest.
// Only URLs under the project's GitHub releases path are allowed.
// progressFn is called periodically with bytes downloaded and total size.
func (c *Client) DownloadAsset(
	ctx context.Context,
	url string,
	dest io.Writer,
	progressFn func(downloaded, total int64),
) error {
	if !strings.HasPrefix(url, allowedDownloadURLBase) {
		return fmt.Errorf("blocked non-release URL: %s", url)
	}

	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.http.Do(req) //nolint:gosec // URL is validated against allowlist above
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	total := resp.ContentLength

	cw := &countingWriter{dest: dest, total: total, progressFn: progressFn}
	if _, err := io.Copy(cw, resp.Body); err != nil {
		return fmt.Errorf("write asset: %w", err)
	}
	return nil
}

// countingWriter wraps an io.Writer and reports progress.
type countingWriter struct {
	dest       io.Writer
	written    int64
	total      int64
	progressFn func(downloaded, total int64)
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.dest.Write(p)
	w.written += int64(n)
	if w.progressFn != nil {
		w.progressFn(w.written, w.total)
	}
	return n, err
}
