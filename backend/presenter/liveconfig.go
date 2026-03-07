package presenter

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/lugassawan/panen/backend/infra/liveconfig"
)

// LiveConfigHandler manages live-reloadable config loaders.
type LiveConfigHandler struct {
	ctx       context.Context
	mu        sync.RWMutex
	loaders   map[string]liveconfig.ConfigLoader
	reloaders map[string]func(ctx context.Context)
}

// Init stores the Wails context for use in handler methods.
func (h *LiveConfigHandler) Init(ctx context.Context) {
	h.ctx = ctx
	h.loaders = make(map[string]liveconfig.ConfigLoader)
	h.reloaders = make(map[string]func(ctx context.Context))
}

// RegisterLoader registers a config loader with a reloader callback.
func (h *LiveConfigHandler) RegisterLoader(
	name string,
	loader liveconfig.ConfigLoader,
	reloader func(ctx context.Context),
) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.loaders[name] = loader
	h.reloaders[name] = reloader
}

// GetAllConfigStatus returns status info for all registered config loaders, sorted by name.
func (h *LiveConfigHandler) GetAllConfigStatus() []*ConfigStatusResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*ConfigStatusResponse, 0, len(h.loaders))
	for _, loader := range h.loaders {
		s := loader.Status()
		result = append(result, &ConfigStatusResponse{
			Name:        s.Name,
			Source:      string(s.Source),
			LastRefresh: s.LastRefresh.Format(dateLayout),
			DataHash:    s.Hash,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ForceRefresh reloads config(s). Pass "" to refresh all, or a name for a specific one.
func (h *LiveConfigHandler) ForceRefresh(configName string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if configName == "" {
		for name := range h.loaders {
			h.refreshOne(name)
		}
		return nil
	}

	if _, ok := h.loaders[configName]; !ok {
		return errors.New("unknown config: " + configName)
	}
	h.refreshOne(configName)
	return nil
}

func (h *LiveConfigHandler) refreshOne(name string) {
	h.loaders[name].Reload(h.ctx)
	if reloader := h.reloaders[name]; reloader != nil {
		reloader(h.ctx)
	}
}
