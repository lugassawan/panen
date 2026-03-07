package liveconfig

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lugassawan/panen/backend/infra/applog"
)

// Source indicates where config data was loaded from.
type Source string

const (
	SourceRemote  Source = "remote"
	SourceCache   Source = "cache"
	SourceBundled Source = "bundled"

	maxBodySize     = 1 << 20 // 1 MB
	refreshInterval = 24 * time.Hour
)

// Config describes a live-reloadable configuration resource.
type Config[T any] struct {
	Name          string
	RemoteURL     string
	CacheFileName string
	BundledData   []byte
	ParseFunc     func([]byte) (T, error)
	ZeroValue     T
}

// Result is the outcome of loading a config.
type Result[T any] struct {
	Data   T
	Source Source
	Hash   string
}

// SettingsStore reads and writes key-value settings.
type SettingsStore interface {
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
}

// EventEmitter sends named events to the frontend.
type EventEmitter interface {
	Emit(eventName string, data any)
}

// StatusInfo describes the current state of a config loader.
type StatusInfo struct {
	Name        string
	Source      Source
	LastRefresh time.Time
	Hash        string
}

// ConfigLoader is a non-generic interface for heterogeneous storage in the presenter.
type ConfigLoader interface {
	Reload(ctx context.Context)
	Status() StatusInfo
}

// Loader fetches, caches, and manages a single live config resource.
type Loader[T any] struct {
	cfg     Config[T]
	deps    Deps
	dataDir string
	client  *http.Client

	mu          sync.RWMutex
	lastResult  Result[T]
	lastRefresh time.Time
}

// Deps holds optional dependencies for change detection and event emission.
type Deps struct {
	Settings SettingsStore
	Emitter  EventEmitter
}

// NewLoader creates a Loader for the given config.
func NewLoader[T any](dataDir string, cfg Config[T], deps Deps) *Loader[T] {
	return &Loader[T]{
		cfg:     cfg,
		deps:    deps,
		dataDir: dataDir,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Load returns the config data using the three-layer fallback: remote -> cache -> bundled.
// Load is safe for concurrent use; only one load executes at a time.
func (l *Loader[T]) Load(ctx context.Context) Result[T] {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.shouldFetchRemote(ctx) {
		if r, ok := l.tryCache(); ok {
			return r
		}
		return l.tryBundled()
	}

	if data, err := l.fetchRemote(ctx); err == nil {
		hash := computeHash(data)
		if parsed, err := l.cfg.ParseFunc(data); err == nil {
			l.writeCache(data)
			l.updateRefreshTimestamp(ctx)
			l.detectChange(ctx, hash)
			r := Result[T]{Data: parsed, Source: SourceRemote, Hash: hash}
			l.lastResult = r
			l.lastRefresh = time.Now()
			return r
		}
	}

	if r, ok := l.tryCache(); ok {
		return r
	}

	return l.tryBundled()
}

// LastResult returns the result from the most recent Load call.
// This avoids redundant loads when the caller already triggered one via Reload.
func (l *Loader[T]) LastResult() Result[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastResult
}

// Reload forces a refresh, bypassing the interval check.
func (l *Loader[T]) Reload(ctx context.Context) {
	if l.deps.Settings != nil {
		_ = l.deps.Settings.SetSetting(ctx, l.refreshKey(), "")
	}
	l.Load(ctx)
}

// Status returns metadata about the last load.
func (l *Loader[T]) Status() StatusInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return StatusInfo{
		Name:        l.cfg.Name,
		Source:      l.lastResult.Source,
		LastRefresh: l.lastRefresh,
		Hash:        l.lastResult.Hash,
	}
}

func (l *Loader[T]) shouldFetchRemote(ctx context.Context) bool {
	if l.deps.Settings == nil {
		return true
	}
	val, err := l.deps.Settings.GetSetting(ctx, l.refreshKey())
	if err != nil || val == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return true
	}
	return time.Since(t) >= refreshInterval
}

func (l *Loader[T]) fetchRemote(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.cfg.RemoteURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := l.client.Do(req) //nolint:gosec // G107: URL is a compile-time constant from Config, not user input
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
}

// tryCache and tryBundled update lastResult/lastRefresh directly (caller holds mu).
func (l *Loader[T]) tryCache() (Result[T], bool) {
	data, err := os.ReadFile(filepath.Join(l.dataDir, l.cfg.CacheFileName))
	if err != nil {
		return Result[T]{}, false
	}
	parsed, err := l.cfg.ParseFunc(data)
	if err != nil {
		return Result[T]{}, false
	}
	r := Result[T]{Data: parsed, Source: SourceCache, Hash: computeHash(data)}
	l.lastResult = r
	l.lastRefresh = time.Now()
	return r, true
}

func (l *Loader[T]) tryBundled() Result[T] {
	parsed, err := l.cfg.ParseFunc(l.cfg.BundledData)
	if err != nil {
		r := Result[T]{Data: l.cfg.ZeroValue, Source: SourceBundled}
		l.lastResult = r
		l.lastRefresh = time.Now()
		return r
	}
	r := Result[T]{Data: parsed, Source: SourceBundled, Hash: computeHash(l.cfg.BundledData)}
	l.lastResult = r
	l.lastRefresh = time.Now()
	return r
}

func (l *Loader[T]) writeCache(data []byte) {
	path := filepath.Join(l.dataDir, l.cfg.CacheFileName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		applog.Warn("write config cache", err, applog.Fields{
			"config": l.cfg.Name,
			"path":   path,
		})
	}
}

func (l *Loader[T]) updateRefreshTimestamp(ctx context.Context) {
	if l.deps.Settings == nil {
		return
	}
	_ = l.deps.Settings.SetSetting(ctx, l.refreshKey(), time.Now().Format(time.RFC3339))
}

func (l *Loader[T]) detectChange(ctx context.Context, newHash string) {
	if l.deps.Settings == nil {
		return
	}
	hashKey := "config_hash_" + l.cfg.Name
	oldHash, _ := l.deps.Settings.GetSetting(ctx, hashKey)
	if oldHash == newHash {
		return
	}
	_ = l.deps.Settings.SetSetting(ctx, hashKey, newHash)

	applog.Info("config changed", applog.Fields{
		"config":  l.cfg.Name,
		"oldHash": oldHash,
		"newHash": newHash,
		"source":  string(SourceRemote),
	})

	if l.deps.Emitter != nil {
		l.deps.Emitter.Emit("config:changed", map[string]string{
			"name": l.cfg.Name,
			"hash": newHash,
		})
	}
}

func (l *Loader[T]) refreshKey() string {
	return "config_last_refresh_" + l.cfg.Name
}

func computeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
