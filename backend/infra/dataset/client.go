package dataset

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lugassawan/panen/backend/domain/dataset"
)

const (
	cacheTTL         = 24 * time.Hour
	httpTimeout      = 10 * time.Second
	maxResponseBytes = 1 << 20 // 1 MB
)

// Client fetches financial datasets from an Apps Script endpoint with local file caching.
type Client struct {
	baseURL  string
	cacheDir string
	client   *http.Client
	mu       sync.Mutex
}

// NewClient creates a Client. When appscriptURL is empty, all methods return nil/empty gracefully.
func NewClient(appscriptURL string, dataDir string) *Client {
	return &Client{
		baseURL:  appscriptURL,
		cacheDir: filepath.Join(dataDir, "datasets"),
		client:   &http.Client{Timeout: httpTimeout},
	}
}

// FetchFinancials returns financials for a single ticker.
func (c *Client) FetchFinancials(ctx context.Context, ticker string) (*dataset.TickerFinancials, error) {
	if c.baseURL == "" {
		return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
	}

	params := url.Values{"action": {"financials"}, "ticker": {ticker}}
	cacheKey := fmt.Sprintf("financials_%s.json", ticker)

	var resp financialsResponse
	if err := c.fetchJSON(ctx, params, cacheKey, &resp); err != nil {
		return nil, err
	}
	return convertTickerFinancials(resp.Data), nil
}

// FetchSectorMetrics returns sector-specific metrics for a ticker.
func (c *Client) FetchSectorMetrics(
	ctx context.Context,
	ticker string,
) (map[string][]dataset.SectorMetricEntry, error) {
	if c.baseURL == "" {
		return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
	}

	params := url.Values{"action": {"sector_metrics"}, "ticker": {ticker}}
	cacheKey := fmt.Sprintf("sector_metrics_%s.json", ticker)

	var resp sectorMetricsResponse
	if err := c.fetchJSON(ctx, params, cacheKey, &resp); err != nil {
		return nil, err
	}
	return convertSectorMetrics(resp.Data), nil
}

// FetchDefinitions returns metric definitions per sector.
func (c *Client) FetchDefinitions(ctx context.Context) (dataset.MetricDefinitions, error) {
	if c.baseURL == "" {
		return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
	}

	params := url.Values{"action": {"definitions"}}
	cacheKey := "definitions.json"

	var resp definitionsResponse
	if err := c.fetchJSON(ctx, params, cacheKey, &resp); err != nil {
		return nil, err
	}
	return dataset.MetricDefinitions(resp.Data), nil
}

// fetchJSON fetches data from the endpoint or cache and decodes into dest.
func (c *Client) fetchJSON(ctx context.Context, params url.Values, cacheKey string, dest any) error {
	cachePath := filepath.Join(c.cacheDir, cacheKey)

	// Check cache.
	if data, err := c.readCache(cachePath); err == nil {
		return json.Unmarshal(data, dest)
	}

	// Fetch from endpoint.
	reqURL := c.baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req) //nolint:gosec // URL is from build-time ldflags, not user input
	if err != nil {
		return fmt.Errorf("fetch dataset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	c.writeCache(cachePath, body)
	return nil
}

// readCache returns cached data if the file exists and is within TTL.
func (c *Client) readCache(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if time.Since(info.ModTime()) > cacheTTL {
		return nil, errors.New("cache expired")
	}
	return os.ReadFile(path)
}

// writeCache atomically writes data to the cache file, creating directories as needed.
func (c *Client) writeCache(path string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return
	}

	// Write to temp file then rename for atomicity — prevents partial reads.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// Wire protocol types for JSON decoding.

type financialsResponse struct {
	Version   int                   `json:"version"`
	UpdatedAt string                `json:"updatedAt"`
	Data      []financialsEntryWire `json:"data"`
}

type sectorMetricsResponse struct {
	Version   int                                `json:"version"`
	UpdatedAt string                             `json:"updatedAt"`
	Data      map[string][]sectorMetricEntryWire `json:"data"`
}

type definitionsResponse struct {
	Version   int                 `json:"version"`
	UpdatedAt string              `json:"updatedAt"`
	Data      map[string][]string `json:"data"`
}

type financialsEntryWire struct {
	Year    int     `json:"year"`
	Quarter int     `json:"quarter"`
	Metric  string  `json:"metric"`
	Value   float64 `json:"value"`
}

type sectorMetricEntryWire struct {
	Year    int     `json:"year"`
	Quarter int     `json:"quarter"`
	Value   float64 `json:"value"`
	Source  string  `json:"source"`
}

// Conversion helpers.

func convertTickerFinancials(entries []financialsEntryWire) *dataset.TickerFinancials {
	if len(entries) == 0 {
		return nil
	}
	tf := &dataset.TickerFinancials{}
	for _, e := range entries {
		ts := dataset.TimeSeriesEntry{Year: e.Year, Quarter: e.Quarter, Value: e.Value}
		switch e.Metric {
		case "revenue":
			tf.Revenue = append(tf.Revenue, ts)
		case "net_income":
			tf.NetIncome = append(tf.NetIncome, ts)
		case "total_assets":
			tf.TotalAssets = append(tf.TotalAssets, ts)
		case "total_equity":
			tf.TotalEquity = append(tf.TotalEquity, ts)
		case "total_debt":
			tf.TotalDebt = append(tf.TotalDebt, ts)
		case "operating_income":
			tf.OperatingIncome = append(tf.OperatingIncome, ts)
		}
	}
	return tf
}

func convertSectorMetrics(data map[string][]sectorMetricEntryWire) map[string][]dataset.SectorMetricEntry {
	if len(data) == 0 {
		return nil
	}
	result := make(map[string][]dataset.SectorMetricEntry, len(data))
	for metric, entries := range data {
		converted := make([]dataset.SectorMetricEntry, len(entries))
		for i, e := range entries {
			converted[i] = dataset.SectorMetricEntry{
				Year:    e.Year,
				Quarter: e.Quarter,
				Value:   e.Value,
				Source:  e.Source,
			}
		}
		result[metric] = converted
	}
	return result
}
