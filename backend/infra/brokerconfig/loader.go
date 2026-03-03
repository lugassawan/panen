package brokerconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lugassawan/panen/configs"

	"github.com/lugassawan/panen/backend/domain/brokerconfig"
)

// remoteURL is a var so tests can override it.
var remoteURL = "https://raw.githubusercontent.com/lugassawan/panen/main/configs/brokers.json"

const (
	cacheFileName = "brokers.json"
	maxBodySize   = 1 << 20 // 1 MB
)

// Loader fetches broker configs using a three-layer fallback: remote → cache → bundled.
type Loader struct {
	dataDir string
	client  *http.Client
}

// NewLoader creates a Loader that caches in dataDir.
func NewLoader(dataDir string) *Loader {
	return &Loader{
		dataDir: dataDir,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Load returns broker configs, trying remote first, then local cache, then bundled fallback.
// It never returns an error — worst case it returns the bundled data.
func (l *Loader) Load(ctx context.Context) []*brokerconfig.BrokerConfig {
	if data, err := l.fetchRemote(ctx); err == nil {
		if cfgs := parseBrokers(data); cfgs != nil {
			l.writeCache(data)
			return cfgs
		}
	}

	if data, err := l.readCache(); err == nil {
		if cfgs := parseBrokers(data); cfgs != nil {
			return cfgs
		}
	}

	if cfgs := parseBrokers(configs.BrokersJSON); cfgs != nil {
		return cfgs
	}

	return nil
}

func (l *Loader) fetchRemote(ctx context.Context) ([]byte, error) {
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

func (l *Loader) readCache() ([]byte, error) {
	return os.ReadFile(filepath.Join(l.dataDir, cacheFileName))
}

func (l *Loader) writeCache(data []byte) {
	_ = os.WriteFile(filepath.Join(l.dataDir, cacheFileName), data, 0o600)
}

func parseBrokers(data []byte) []*brokerconfig.BrokerConfig {
	var cfgs []*brokerconfig.BrokerConfig
	if err := json.Unmarshal(data, &cfgs); err != nil {
		return nil
	}
	if len(cfgs) == 0 {
		return nil
	}
	return cfgs
}
