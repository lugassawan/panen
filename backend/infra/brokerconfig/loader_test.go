package brokerconfig

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const validJSON = `[{"code":"XC","name":"Ajaib","buyFeePct":0.15,"sellFeePct":0.15,"sellTaxPct":0.10,"notes":""}]`

func TestLoaderRemoteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(validJSON))
	}))
	defer srv.Close()

	origURL := remoteURL
	remoteURL = srv.URL
	defer func() { remoteURL = origURL }()

	dir := t.TempDir()
	l := NewLoader(dir)

	cfgs := l.Load(context.Background())
	if cfgs == nil {
		t.Fatal("Load() returned nil for valid remote")
	}
	if len(cfgs) != 1 {
		t.Fatalf("len = %d, want 1", len(cfgs))
	}
	if cfgs[0].Code != "XC" {
		t.Errorf("Code = %q, want XC", cfgs[0].Code)
	}
	if cfgs[0].BuyFeePct != 0.15 {
		t.Errorf("BuyFeePct = %v, want 0.15", cfgs[0].BuyFeePct)
	}

	// Verify cache was written
	data, err := os.ReadFile(filepath.Join(dir, cacheFileName))
	if err != nil {
		t.Fatalf("cache not written: %v", err)
	}
	if string(data) != validJSON {
		t.Errorf("cached data = %q, want %q", string(data), validJSON)
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
	l := NewLoader(dir)

	// Remote returns 404 — Load should fall back to bundled
	cfgs := l.Load(context.Background())
	if cfgs == nil {
		t.Fatal("Load() returned nil — bundled fallback should have kicked in")
	}
}

func TestLoaderCacheFallback(t *testing.T) {
	dir := t.TempDir()

	// Write cached data
	cachePath := filepath.Join(dir, cacheFileName)
	if err := os.WriteFile(cachePath, []byte(validJSON), 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	l := NewLoader(dir)

	data, err := l.readCache()
	if err != nil {
		t.Fatalf("readCache() error = %v", err)
	}

	cfgs := parseBrokers(data)
	if cfgs == nil {
		t.Fatal("parseBrokers returned nil for cached data")
	}
	if cfgs[0].Code != "XC" {
		t.Errorf("Code = %q, want XC", cfgs[0].Code)
	}
}

func TestLoaderBundledFallback(t *testing.T) {
	dir := t.TempDir()
	l := NewLoader(dir)

	// No cache, remote will fail (no server) — Load should use bundled
	cfgs := l.Load(context.Background())
	if cfgs == nil {
		t.Fatal("Load() returned nil — bundled fallback failed")
	}
	if len(cfgs) == 0 {
		t.Fatal("Load() returned empty slice")
	}
	// Verify at least one known broker from bundled data
	found := false
	for _, c := range cfgs {
		if c.Code == "XC" {
			found = true
			break
		}
	}
	if !found {
		t.Error("bundled data does not contain expected broker code XC")
	}
}

func TestLoaderWriteCache(t *testing.T) {
	dir := t.TempDir()
	l := NewLoader(dir)

	l.writeCache([]byte(validJSON))

	data, err := os.ReadFile(filepath.Join(dir, cacheFileName))
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if string(data) != validJSON {
		t.Errorf("cached data = %q, want %q", string(data), validJSON)
	}
}

func TestParseBrokersMalformedJSON(t *testing.T) {
	cfgs := parseBrokers([]byte("not json"))
	if cfgs != nil {
		t.Errorf("expected nil for malformed JSON, got %v", cfgs)
	}
}

func TestParseBrokersEmptyArray(t *testing.T) {
	cfgs := parseBrokers([]byte("[]"))
	if cfgs != nil {
		t.Errorf("expected nil for empty array, got %v", cfgs)
	}
}
