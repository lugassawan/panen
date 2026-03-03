package watchlistconfig

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

const validIndicesJSON = `{"IDX30":["BBCA","BBRI","BMRI"],"LQ45":["BBCA","BBRI","BMRI","TLKM"]}`

func TestLoaderRemoteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(validIndicesJSON))
	}))
	defer srv.Close()

	origURL := remoteURL
	remoteURL = srv.URL
	defer func() { remoteURL = origURL }()

	dir := t.TempDir()
	l := NewIndexLoader(dir)

	reg := l.Load(context.Background())
	if reg == nil {
		t.Fatal("Load() returned nil for valid remote")
	}

	tickers, ok := reg.Tickers("IDX30")
	if !ok {
		t.Fatal("Tickers(IDX30) not found")
	}
	if len(tickers) != 3 {
		t.Fatalf("len(tickers) = %d, want 3", len(tickers))
	}

	// Verify cache was written
	data, err := os.ReadFile(filepath.Join(dir, cacheFileName))
	if err != nil {
		t.Fatalf("cache not written: %v", err)
	}
	if string(data) != validIndicesJSON {
		t.Errorf("cached data = %q, want %q", string(data), validIndicesJSON)
	}
}

func TestLoaderRemoteNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	origURL := remoteURL
	remoteURL = srv.URL
	defer func() { remoteURL = origURL }()

	dir := t.TempDir()
	l := NewIndexLoader(dir)

	// Remote returns 404 — Load should fall back to bundled
	reg := l.Load(context.Background())
	if reg == nil {
		t.Fatal("Load() returned nil — bundled fallback should have kicked in")
	}
}

func TestLoaderCacheFallback(t *testing.T) {
	// Use an unreachable host to force remote failure
	origURL := remoteURL
	remoteURL = "http://127.0.0.1:0/indices.json"
	defer func() { remoteURL = origURL }()

	dir := t.TempDir()

	// Write cached data
	cachePath := filepath.Join(dir, cacheFileName)
	if err := os.WriteFile(cachePath, []byte(validIndicesJSON), 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	l := NewIndexLoader(dir)
	reg := l.Load(context.Background())
	if reg == nil {
		t.Fatal("Load() returned nil — cache fallback should have kicked in")
	}

	tickers, ok := reg.Tickers("LQ45")
	if !ok {
		t.Fatal("Tickers(LQ45) not found in cached data")
	}
	if len(tickers) != 4 {
		t.Fatalf("len(tickers) = %d, want 4", len(tickers))
	}
}

func TestLoaderBundledFallback(t *testing.T) {
	dir := t.TempDir()
	l := NewIndexLoader(dir)

	// No cache, remote will fail (no server) — Load should use bundled
	reg := l.Load(context.Background())
	if reg == nil {
		t.Fatal("Load() returned nil — bundled fallback failed")
	}

	names := reg.Names()
	if len(names) == 0 {
		t.Fatal("Names() returned empty slice from bundled data")
	}

	// Verify at least one known index from bundled data
	_, ok := reg.Tickers("IDX30")
	if !ok {
		t.Error("bundled data does not contain expected index IDX30")
	}
}

func TestLoaderWriteCache(t *testing.T) {
	dir := t.TempDir()
	l := NewIndexLoader(dir)

	l.writeCache([]byte(validIndicesJSON))

	data, err := os.ReadFile(filepath.Join(dir, cacheFileName))
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if string(data) != validIndicesJSON {
		t.Errorf("cached data = %q, want %q", string(data), validIndicesJSON)
	}
}

func TestParseIndicesMalformedJSON(t *testing.T) {
	reg := parseIndices([]byte("not json"))
	if reg != nil {
		t.Errorf("expected nil for malformed JSON, got %v", reg)
	}
}

func TestParseIndicesEmptyObject(t *testing.T) {
	reg := parseIndices([]byte("{}"))
	if reg != nil {
		t.Errorf("expected nil for empty object, got %v", reg)
	}
}

func TestIndexRegistryKnownIndex(t *testing.T) {
	reg := parseIndices([]byte(validIndicesJSON))
	if reg == nil {
		t.Fatal("parseIndices returned nil")
	}

	tickers, ok := reg.Tickers("IDX30")
	if !ok {
		t.Fatal("Tickers(IDX30) not found")
	}
	if len(tickers) != 3 {
		t.Fatalf("len = %d, want 3", len(tickers))
	}
	if tickers[0] != "BBCA" {
		t.Errorf("tickers[0] = %q, want BBCA", tickers[0])
	}
}

func TestIndexRegistryUnknownIndex(t *testing.T) {
	reg := parseIndices([]byte(validIndicesJSON))
	if reg == nil {
		t.Fatal("parseIndices returned nil")
	}

	tickers, ok := reg.Tickers("NONEXISTENT")
	if ok {
		t.Error("Tickers(NONEXISTENT) should return ok=false")
	}
	if tickers != nil {
		t.Errorf("Tickers(NONEXISTENT) = %v, want nil", tickers)
	}
}

func TestIndexRegistryNames(t *testing.T) {
	reg := parseIndices([]byte(validIndicesJSON))
	if reg == nil {
		t.Fatal("parseIndices returned nil")
	}

	names := reg.Names()
	if len(names) != 2 {
		t.Fatalf("len(Names()) = %d, want 2", len(names))
	}
	// Names() must be sorted
	if !sort.StringsAreSorted(names) {
		t.Errorf("Names() = %v is not sorted", names)
	}
	if names[0] != "IDX30" {
		t.Errorf("names[0] = %q, want IDX30", names[0])
	}
	if names[1] != "LQ45" {
		t.Errorf("names[1] = %q, want LQ45", names[1])
	}
}

func TestSectorRegistryKnownTicker(t *testing.T) {
	sr := NewSectorRegistry()

	tests := []struct {
		ticker string
		want   string
	}{
		{"BBCA", "Banking"},
		{"TLKM", "Telco"},
		{"GOTO", "Technology"},
		{"ADRO", "Mining"},
	}

	for _, tt := range tests {
		t.Run(tt.ticker, func(t *testing.T) {
			got := sr.SectorOf(tt.ticker)
			if got != tt.want {
				t.Errorf("SectorOf(%q) = %q, want %q", tt.ticker, got, tt.want)
			}
		})
	}
}

func TestSectorRegistryUnknownTicker(t *testing.T) {
	sr := NewSectorRegistry()

	got := sr.SectorOf("UNKNOWN")
	if got != "" {
		t.Errorf("SectorOf(UNKNOWN) = %q, want empty string", got)
	}
}

func TestSectorRegistryAllSectors(t *testing.T) {
	sr := NewSectorRegistry()

	sectors := sr.AllSectors()
	if len(sectors) == 0 {
		t.Fatal("AllSectors() returned empty slice")
	}

	// Must be sorted
	if !sort.StringsAreSorted(sectors) {
		t.Errorf("AllSectors() = %v is not sorted", sectors)
	}

	// Must be unique
	seen := make(map[string]struct{}, len(sectors))
	for _, s := range sectors {
		if _, dup := seen[s]; dup {
			t.Errorf("AllSectors() contains duplicate: %q", s)
		}
		seen[s] = struct{}{}
	}

	// Verify known sectors are present
	knownSectors := []string{"Banking", "Mining", "Technology", "Telco", "Consumer", "Energy"}
	for _, want := range knownSectors {
		if _, ok := seen[want]; !ok {
			t.Errorf("AllSectors() missing expected sector %q", want)
		}
	}
}
