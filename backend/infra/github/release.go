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
	defaultAPIURL   = "https://api.github.com"
	defaultRepo     = "lugassawan/panen"
	maxResponseSize = 1 << 20 // 1 MB
)

// Release holds the relevant fields from a GitHub release response.
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
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
