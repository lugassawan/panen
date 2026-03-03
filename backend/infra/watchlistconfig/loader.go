package watchlistconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/lugassawan/panen/configs"
)

// remoteURL is a var so tests can override it.
var remoteURL = "https://raw.githubusercontent.com/lugassawan/panen/main/configs/indices.json"

const (
	cacheFileName = "indices.json"
	maxBodySize   = 1 << 20 // 1 MB
)

// IndexLoader fetches index compositions using a three-layer fallback: remote → cache → bundled.
type IndexLoader struct {
	dataDir string
	client  *http.Client
}

// NewIndexLoader creates an IndexLoader that caches in dataDir.
func NewIndexLoader(dataDir string) *IndexLoader {
	return &IndexLoader{
		dataDir: dataDir,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Load returns an IndexRegistry, trying remote first, then local cache, then bundled fallback.
// It never returns an error — worst case it returns the bundled data.
func (l *IndexLoader) Load(ctx context.Context) *IndexRegistry {
	if data, err := l.fetchRemote(ctx); err == nil {
		if reg := parseIndices(data); reg != nil {
			l.writeCache(data)
			return reg
		}
	}

	if data, err := l.readCache(); err == nil {
		if reg := parseIndices(data); reg != nil {
			return reg
		}
	}

	if reg := parseIndices(configs.IndicesJSON); reg != nil {
		return reg
	}

	return &IndexRegistry{indices: map[string][]string{}}
}

func (l *IndexLoader) fetchRemote(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := l.client.Do(req) //nolint:gosec // URL is a compile-time constant
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
}

func (l *IndexLoader) readCache() ([]byte, error) {
	return os.ReadFile(filepath.Join(l.dataDir, cacheFileName))
}

func (l *IndexLoader) writeCache(data []byte) {
	_ = os.WriteFile(filepath.Join(l.dataDir, cacheFileName), data, 0o600)
}

// IndexRegistry provides lookup access to index compositions.
type IndexRegistry struct {
	indices map[string][]string
}

// Tickers returns the list of tickers for the named index, and whether it was found.
func (r *IndexRegistry) Tickers(name string) ([]string, bool) {
	tickers, ok := r.indices[name]
	return tickers, ok
}

// Names returns a sorted list of all index names.
func (r *IndexRegistry) Names() []string {
	names := make([]string, 0, len(r.indices))
	for name := range r.indices {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SectorRegistry maps tickers to their sector. Bundled-only (sectors are stable).
type SectorRegistry struct {
	sectors map[string]string
}

// NewSectorRegistry parses the bundled sectors config and returns a SectorRegistry.
func NewSectorRegistry() *SectorRegistry {
	var m map[string]string
	if err := json.Unmarshal(configs.SectorsJSON, &m); err != nil {
		return &SectorRegistry{sectors: map[string]string{}}
	}
	return &SectorRegistry{sectors: m}
}

// SectorOf returns the sector for the given ticker, or "" if unknown.
func (r *SectorRegistry) SectorOf(ticker string) string {
	return r.sectors[ticker]
}

// AllSectors returns a sorted list of unique sector names.
func (r *SectorRegistry) AllSectors() []string {
	seen := make(map[string]struct{}, len(r.sectors))
	for _, sector := range r.sectors {
		seen[sector] = struct{}{}
	}
	sectors := make([]string, 0, len(seen))
	for sector := range seen {
		sectors = append(sectors, sector)
	}
	sort.Strings(sectors)
	return sectors
}

func parseIndices(data []byte) *IndexRegistry {
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	if len(m) == 0 {
		return nil
	}
	return &IndexRegistry{indices: m}
}
