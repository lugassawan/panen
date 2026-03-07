package liveconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type mockSettings struct {
	mu   sync.Mutex
	data map[string]string
}

func newMockSettings() *mockSettings {
	return &mockSettings{data: make(map[string]string)}
}

func (m *mockSettings) GetSetting(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key], nil
}

func (m *mockSettings) SetSetting(_ context.Context, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

type mockEmitter struct {
	mu     sync.Mutex
	events []emittedEvent
}

type emittedEvent struct {
	Name string
	Data any
}

func (m *mockEmitter) Emit(eventName string, data any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, emittedEvent{Name: eventName, Data: data})
}

func (m *mockEmitter) eventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func testConfig(url string) Config[[]string] {
	return Config[[]string]{
		Name:          "test",
		RemoteURL:     url,
		CacheFileName: "test.json",
		BundledData:   []byte(`["bundled"]`),
		ParseFunc: func(data []byte) ([]string, error) {
			var v []string
			err := json.Unmarshal(data, &v)
			return v, err
		},
	}
}

func TestRemoteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`["remote"]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	settings := newMockSettings()
	emitter := &mockEmitter{}
	l := NewLoader(dir, testConfig(srv.URL), Deps{Settings: settings, Emitter: emitter})

	r := l.Load(context.Background())
	if r.Source != SourceRemote {
		t.Errorf("Source = %q, want %q", r.Source, SourceRemote)
	}
	if len(r.Data) != 1 || r.Data[0] != "remote" {
		t.Errorf("Data = %v, want [remote]", r.Data)
	}
	if r.Hash == "" {
		t.Error("Hash is empty")
	}

	// Verify cache was written
	data, err := os.ReadFile(filepath.Join(dir, "test.json"))
	if err != nil {
		t.Fatalf("cache not written: %v", err)
	}
	if string(data) != `["remote"]` {
		t.Errorf("cached = %q, want %q", string(data), `["remote"]`)
	}
}

func TestRemoteFailureCacheFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.json"), []byte(`["cached"]`), 0o600); err != nil {
		t.Fatal(err)
	}

	l := NewLoader(dir, testConfig(srv.URL), Deps{})
	r := l.Load(context.Background())

	if r.Source != SourceCache {
		t.Errorf("Source = %q, want %q", r.Source, SourceCache)
	}
	if len(r.Data) != 1 || r.Data[0] != "cached" {
		t.Errorf("Data = %v, want [cached]", r.Data)
	}
}

func TestCacheFailureBundledFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := t.TempDir()
	l := NewLoader(dir, testConfig(srv.URL), Deps{})
	r := l.Load(context.Background())

	if r.Source != SourceBundled {
		t.Errorf("Source = %q, want %q", r.Source, SourceBundled)
	}
	if len(r.Data) != 1 || r.Data[0] != "bundled" {
		t.Errorf("Data = %v, want [bundled]", r.Data)
	}
}

func TestTotalFailureReturnsZeroValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := t.TempDir()
	cfg := testConfig(srv.URL)
	cfg.BundledData = []byte("invalid json")
	cfg.ZeroValue = []string{"zero"}

	l := NewLoader(dir, cfg, Deps{})
	r := l.Load(context.Background())

	if r.Source != SourceBundled {
		t.Errorf("Source = %q, want %q", r.Source, SourceBundled)
	}
	if len(r.Data) != 1 || r.Data[0] != "zero" {
		t.Errorf("Data = %v, want [zero]", r.Data)
	}
}

func TestChangeDetectionEmitsEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`["v1"]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	settings := newMockSettings()
	emitter := &mockEmitter{}
	l := NewLoader(dir, testConfig(srv.URL), Deps{Settings: settings, Emitter: emitter})

	// First load — no previous hash, should emit change
	l.Load(context.Background())
	if emitter.eventCount() != 1 {
		t.Errorf("events after first load = %d, want 1", emitter.eventCount())
	}

	// Second load with same data — should NOT emit
	// Clear refresh timestamp to force remote fetch
	_ = settings.SetSetting(context.Background(), "config_last_refresh_test", "")
	l.Load(context.Background())
	if emitter.eventCount() != 1 {
		t.Errorf("events after second load = %d, want 1", emitter.eventCount())
	}
}

func TestChangeDetectionNewHashEmitsAgain(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 1 {
			_, _ = w.Write([]byte(`["v1"]`))
		} else {
			_, _ = w.Write([]byte(`["v2"]`))
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	settings := newMockSettings()
	emitter := &mockEmitter{}
	l := NewLoader(dir, testConfig(srv.URL), Deps{Settings: settings, Emitter: emitter})

	l.Load(context.Background())
	// Clear refresh timestamp to force second remote fetch
	_ = settings.SetSetting(context.Background(), "config_last_refresh_test", "")
	l.Load(context.Background())

	if emitter.eventCount() != 2 {
		t.Errorf("events = %d, want 2", emitter.eventCount())
	}
}

func TestRefreshIntervalSkipsRemote(t *testing.T) {
	remoteHit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		remoteHit = true
		_, _ = w.Write([]byte(`["remote"]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	settings := newMockSettings()
	// Set recent timestamp
	_ = settings.SetSetting(context.Background(), "config_last_refresh_test", time.Now().Format(time.RFC3339))

	// Write cache so fallback works
	if err := os.WriteFile(filepath.Join(dir, "test.json"), []byte(`["cached"]`), 0o600); err != nil {
		t.Fatal(err)
	}

	l := NewLoader(dir, testConfig(srv.URL), Deps{Settings: settings})
	r := l.Load(context.Background())

	if remoteHit {
		t.Error("remote was fetched despite recent refresh timestamp")
	}
	if r.Source != SourceCache {
		t.Errorf("Source = %q, want %q", r.Source, SourceCache)
	}
}

func TestRefreshIntervalExpiredFetchesRemote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`["remote"]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	settings := newMockSettings()
	// Set old timestamp (>24h ago)
	old := time.Now().Add(-25 * time.Hour).Format(time.RFC3339)
	_ = settings.SetSetting(context.Background(), "config_last_refresh_test", old)

	l := NewLoader(dir, testConfig(srv.URL), Deps{Settings: settings})
	r := l.Load(context.Background())

	if r.Source != SourceRemote {
		t.Errorf("Source = %q, want %q", r.Source, SourceRemote)
	}
}

func TestReloadBypassesInterval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`["remote"]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	settings := newMockSettings()
	// Set recent timestamp
	_ = settings.SetSetting(context.Background(), "config_last_refresh_test", time.Now().Format(time.RFC3339))

	l := NewLoader(dir, testConfig(srv.URL), Deps{Settings: settings})
	l.Reload(context.Background())

	s := l.Status()
	if s.Source != SourceRemote {
		t.Errorf("Source after Reload = %q, want %q", s.Source, SourceRemote)
	}
}

func TestStatusReturnsMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`["data"]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	l := NewLoader(dir, testConfig(srv.URL), Deps{})
	l.Load(context.Background())

	s := l.Status()
	if s.Name != "test" {
		t.Errorf("Name = %q, want test", s.Name)
	}
	if s.Source != SourceRemote {
		t.Errorf("Source = %q, want %q", s.Source, SourceRemote)
	}
	if s.Hash == "" {
		t.Error("Hash is empty")
	}
	if s.LastRefresh.IsZero() {
		t.Error("LastRefresh is zero")
	}
}

func TestNilDepsNoPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`["data"]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	l := NewLoader(dir, testConfig(srv.URL), Deps{})

	// Should not panic with nil Settings and nil Emitter
	r := l.Load(context.Background())
	if r.Source != SourceRemote {
		t.Errorf("Source = %q, want %q", r.Source, SourceRemote)
	}

	l.Reload(context.Background())
	_ = l.Status()
}
