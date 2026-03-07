package watchlistconfig

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/lugassawan/panen/backend/infra/liveconfig"
)

const validIndicesJSON = `{"IDX30":["BBCA","BBRI","BMRI"],"LQ45":["BBCA","BBRI","BMRI","TLKM"]}`

func testIndexLoader(dir, url string) *liveconfig.Loader[*IndexRegistry] {
	l := NewIndexLoader(dir, liveconfig.Deps{})
	l.SetRemoteURL(url)
	return l
}

func TestLoaderRemoteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(validIndicesJSON))
	}))
	defer srv.Close()

	dir := t.TempDir()
	l := testIndexLoader(dir, srv.URL)
	r := l.Load(context.Background())

	if r.Source != liveconfig.SourceRemote {
		t.Fatalf("Source = %q, want %q", r.Source, liveconfig.SourceRemote)
	}

	tickers, ok := r.Data.Tickers("IDX30")
	if !ok {
		t.Fatal("Tickers(IDX30) not found")
	}
	if len(tickers) != 3 {
		t.Fatalf("len(tickers) = %d, want 3", len(tickers))
	}

	data, err := os.ReadFile(filepath.Join(dir, "indices.json"))
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

	dir := t.TempDir()
	l := testIndexLoader(dir, srv.URL)
	r := l.Load(context.Background())

	if r.Data == nil {
		t.Fatal("Load() returned nil — bundled fallback should have kicked in")
	}
}

func TestLoaderCacheFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := t.TempDir()
	cachePath := filepath.Join(dir, "indices.json")
	if err := os.WriteFile(cachePath, []byte(validIndicesJSON), 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	l := testIndexLoader(dir, srv.URL)
	r := l.Load(context.Background())

	if r.Source != liveconfig.SourceCache {
		t.Errorf("Source = %q, want %q", r.Source, liveconfig.SourceCache)
	}

	tickers, ok := r.Data.Tickers("LQ45")
	if !ok {
		t.Fatal("Tickers(LQ45) not found in cached data")
	}
	if len(tickers) != 4 {
		t.Fatalf("len(tickers) = %d, want 4", len(tickers))
	}
}

func TestLoaderBundledFallback(t *testing.T) {
	dir := t.TempDir()
	l := testIndexLoader(dir, "http://127.0.0.1:0/invalid")
	r := l.Load(context.Background())

	if r.Data == nil {
		t.Fatal("Load() returned nil — bundled fallback failed")
	}

	names := r.Data.Names()
	if len(names) == 0 {
		t.Fatal("Names() returned empty slice from bundled data")
	}

	_, ok := r.Data.Tickers("IDX30")
	if !ok {
		t.Error("bundled data does not contain expected index IDX30")
	}
}

func TestParseIndicesMalformedJSON(t *testing.T) {
	_, err := parseIndices([]byte("not json"))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseIndicesEmptyObject(t *testing.T) {
	_, err := parseIndices([]byte("{}"))
	if err == nil {
		t.Error("expected error for empty object")
	}
}

func TestIndexRegistryKnownIndex(t *testing.T) {
	reg, err := parseIndices([]byte(validIndicesJSON))
	if err != nil {
		t.Fatal(err)
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
	reg, err := parseIndices([]byte(validIndicesJSON))
	if err != nil {
		t.Fatal(err)
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
	reg, err := parseIndices([]byte(validIndicesJSON))
	if err != nil {
		t.Fatal(err)
	}

	names := reg.Names()
	if len(names) != 2 {
		t.Fatalf("len(Names()) = %d, want 2", len(names))
	}
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

func TestSwappableIndexRegistryDelegates(t *testing.T) {
	reg, _ := parseIndices([]byte(validIndicesJSON))
	s := NewSwappableIndexRegistry(reg)

	tickers, ok := s.Tickers("IDX30")
	if !ok {
		t.Fatal("Tickers(IDX30) not found")
	}
	if len(tickers) != 3 {
		t.Fatalf("len = %d, want 3", len(tickers))
	}

	names := s.Names()
	if len(names) != 2 {
		t.Fatalf("len(Names()) = %d, want 2", len(names))
	}
}

func TestSwappableIndexRegistrySwap(t *testing.T) {
	reg1, _ := parseIndices([]byte(`{"A":["X"]}`))
	reg2, _ := parseIndices([]byte(`{"B":["Y","Z"]}`))

	s := NewSwappableIndexRegistry(reg1)

	_, ok := s.Tickers("A")
	if !ok {
		t.Fatal("before swap: Tickers(A) not found")
	}

	s.Swap(reg2)

	_, ok = s.Tickers("A")
	if ok {
		t.Error("after swap: Tickers(A) should not be found")
	}
	tickers, ok := s.Tickers("B")
	if !ok {
		t.Fatal("after swap: Tickers(B) not found")
	}
	if len(tickers) != 2 {
		t.Errorf("len(tickers) = %d, want 2", len(tickers))
	}
}

func TestSwappableIndexRegistryConcurrent(t *testing.T) {
	reg1, _ := parseIndices([]byte(`{"A":["X"]}`))
	reg2, _ := parseIndices([]byte(`{"B":["Y"]}`))
	s := NewSwappableIndexRegistry(reg1)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.Swap(reg2)
		}()
		go func() {
			defer wg.Done()
			_ = s.Names()
			s.Tickers("A")
		}()
	}
	wg.Wait()
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

	if !sort.StringsAreSorted(sectors) {
		t.Errorf("AllSectors() = %v is not sorted", sectors)
	}

	seen := make(map[string]struct{}, len(sectors))
	for _, s := range sectors {
		if _, dup := seen[s]; dup {
			t.Errorf("AllSectors() contains duplicate: %q", s)
		}
		seen[s] = struct{}{}
	}

	knownSectors := []string{"Banking", "Mining", "Technology", "Telco", "Consumer", "Energy"}
	for _, want := range knownSectors {
		if _, ok := seen[want]; !ok {
			t.Errorf("AllSectors() missing expected sector %q", want)
		}
	}
}
