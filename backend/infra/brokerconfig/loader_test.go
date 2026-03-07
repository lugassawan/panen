package brokerconfig

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/lugassawan/panen/backend/domain/brokerconfig"
	"github.com/lugassawan/panen/backend/infra/liveconfig"
	"github.com/lugassawan/panen/configs"
)

const validJSON = `[{"code":"XC","name":"Ajaib","buyFeePct":0.15,"sellFeePct":0.15,"sellTaxPct":0.10,"notes":""}]`

func testLoader(dir, url string) *liveconfig.Loader[[]*brokerconfig.BrokerConfig] {
	return liveconfig.NewLoader(dir, liveconfig.Config[[]*brokerconfig.BrokerConfig]{
		Name:          "brokers",
		RemoteURL:     url,
		CacheFileName: "brokers.json",
		BundledData:   configs.BrokersJSON,
		ParseFunc:     parseBrokers,
	}, liveconfig.Deps{})
}

func TestLoaderRemoteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(validJSON))
	}))
	defer srv.Close()

	dir := t.TempDir()
	l := testLoader(dir, srv.URL)
	r := l.Load(context.Background())

	if r.Source != liveconfig.SourceRemote {
		t.Fatalf("Source = %q, want %q", r.Source, liveconfig.SourceRemote)
	}
	if len(r.Data) != 1 {
		t.Fatalf("len = %d, want 1", len(r.Data))
	}
	if r.Data[0].Code != "XC" {
		t.Errorf("Code = %q, want XC", r.Data[0].Code)
	}
	if r.Data[0].BuyFeePct != 0.15 {
		t.Errorf("BuyFeePct = %v, want 0.15", r.Data[0].BuyFeePct)
	}

	data, err := os.ReadFile(filepath.Join(dir, "brokers.json"))
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

	dir := t.TempDir()
	l := testLoader(dir, srv.URL)
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
	cachePath := filepath.Join(dir, "brokers.json")
	if err := os.WriteFile(cachePath, []byte(validJSON), 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	l := testLoader(dir, srv.URL)
	r := l.Load(context.Background())

	if r.Source != liveconfig.SourceCache {
		t.Errorf("Source = %q, want %q", r.Source, liveconfig.SourceCache)
	}
	if r.Data[0].Code != "XC" {
		t.Errorf("Code = %q, want XC", r.Data[0].Code)
	}
}

func TestLoaderBundledFallback(t *testing.T) {
	dir := t.TempDir()
	l := testLoader(dir, "http://127.0.0.1:0/invalid")
	r := l.Load(context.Background())

	if r.Data == nil {
		t.Fatal("Load() returned nil — bundled fallback failed")
	}
	if len(r.Data) == 0 {
		t.Fatal("Load() returned empty slice")
	}
	found := false
	for _, c := range r.Data {
		if c.Code == "XC" {
			found = true
			break
		}
	}
	if !found {
		t.Error("bundled data does not contain expected broker code XC")
	}
}

func TestParseBrokersMalformedJSON(t *testing.T) {
	_, err := parseBrokers([]byte("not json"))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseBrokersEmptyArray(t *testing.T) {
	_, err := parseBrokers([]byte("[]"))
	if err == nil {
		t.Error("expected error for empty array")
	}
}
