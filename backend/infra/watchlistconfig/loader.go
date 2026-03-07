package watchlistconfig

import (
	"encoding/json"
	"errors"
	"sort"
	"sync"

	"github.com/lugassawan/panen/backend/infra/liveconfig"
	"github.com/lugassawan/panen/configs"
)

const remoteURL = "https://raw.githubusercontent.com/lugassawan/panen/main/configs/indices.json"

// NewIndexLoader creates a liveconfig.Loader for index compositions.
func NewIndexLoader(dataDir string, deps liveconfig.Deps) *liveconfig.Loader[*IndexRegistry] {
	return liveconfig.NewLoader(dataDir, liveconfig.Config[*IndexRegistry]{
		Name:          "indices",
		RemoteURL:     remoteURL,
		CacheFileName: "indices.json",
		BundledData:   configs.IndicesJSON,
		ParseFunc:     parseIndices,
		ZeroValue:     &IndexRegistry{indices: map[string][]string{}},
	}, deps)
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

// SwappableIndexRegistry wraps an IndexRegistry with atomic swap support.
// It implements usecase.IndexRegistry so it can be passed to services directly.
type SwappableIndexRegistry struct {
	mu  sync.RWMutex
	reg *IndexRegistry
}

// NewSwappableIndexRegistry creates a SwappableIndexRegistry with an initial registry.
func NewSwappableIndexRegistry(reg *IndexRegistry) *SwappableIndexRegistry {
	return &SwappableIndexRegistry{reg: reg}
}

// Tickers delegates to the underlying IndexRegistry under a read lock.
func (s *SwappableIndexRegistry) Tickers(name string) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reg.Tickers(name)
}

// Names delegates to the underlying IndexRegistry under a read lock.
func (s *SwappableIndexRegistry) Names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reg.Names()
}

// Swap replaces the underlying IndexRegistry.
func (s *SwappableIndexRegistry) Swap(reg *IndexRegistry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reg = reg
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

func parseIndices(data []byte) (*IndexRegistry, error) {
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, errors.New("empty index config")
	}
	return &IndexRegistry{indices: m}, nil
}
