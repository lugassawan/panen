package presenter

import (
	"context"

	domainProvider "github.com/lugassawan/panen/backend/domain/provider"
)

// ProviderHandler handles provider-related requests from the frontend.
type ProviderHandler struct {
	ctx      context.Context
	registry domainProvider.Registry
}

// Bind wires the handler to its dependencies.
func (h *ProviderHandler) Bind(ctx context.Context, registry domainProvider.Registry) {
	h.ctx = ctx
	h.registry = registry
}

// GetProviderStatus returns the status of all registered data providers.
func (h *ProviderHandler) GetProviderStatus() []ProviderStatusResponse {
	infos := h.registry.List()
	result := make([]ProviderStatusResponse, len(infos))
	for i, info := range infos {
		result[i] = ProviderStatusResponse{
			Name:      info.Name,
			Priority:  info.Priority,
			Status:    string(info.Status),
			LastCheck: formatDTO(info.LastCheck),
			LastError: info.LastError,
			Enabled:   info.Enabled,
		}
	}
	return result
}

// SetProviderEnabled enables or disables a provider by name.
func (h *ProviderHandler) SetProviderEnabled(name string, enabled bool) bool {
	return h.registry.SetEnabled(name, enabled)
}

// RunProviderHealthCheck triggers a health check on all providers and waits for completion.
func (h *ProviderHandler) RunProviderHealthCheck() {
	h.registry.HealthCheckAll(h.ctx)
}
