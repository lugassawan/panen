package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/domain/settings"
)

const (
	settingFMPAPIKey     = "fmp_api_key"     //nolint:gosec // settings key name, not a credential
	settingSectorsAPIKey = "sectors_api_key" //nolint:gosec // settings key name, not a credential
)

// SettingsHandler handles API key settings requests from the frontend.
type SettingsHandler struct {
	ctx      context.Context
	settings settings.Repository
}

// Bind wires the handler to its dependencies.
func (h *SettingsHandler) Bind(ctx context.Context, settings settings.Repository) {
	h.ctx = ctx
	h.settings = settings
}

// GetAPIKeyStatus returns which API keys are configured (never exposes raw keys).
func (h *SettingsHandler) GetAPIKeyStatus() (*APIKeyStatusResponse, error) {
	fmpKey, err := h.settings.GetSetting(h.ctx, settingFMPAPIKey)
	if err != nil {
		return nil, err
	}

	sectorsKey, err := h.settings.GetSetting(h.ctx, settingSectorsAPIKey)
	if err != nil {
		return nil, err
	}

	return &APIKeyStatusResponse{
		FMPConfigured:     fmpKey != "",
		SectorsConfigured: sectorsKey != "",
	}, nil
}

// SetFMPAPIKey stores the FMP API key.
func (h *SettingsHandler) SetFMPAPIKey(key string) error {
	return h.settings.SetSetting(h.ctx, settingFMPAPIKey, key)
}

// SetSectorsAPIKey stores the Sectors.app API key.
func (h *SettingsHandler) SetSectorsAPIKey(key string) error {
	return h.settings.SetSetting(h.ctx, settingSectorsAPIKey, key)
}

// ClearFMPAPIKey removes the FMP API key.
func (h *SettingsHandler) ClearFMPAPIKey() error {
	return h.settings.SetSetting(h.ctx, settingFMPAPIKey, "")
}

// ClearSectorsAPIKey removes the Sectors.app API key.
func (h *SettingsHandler) ClearSectorsAPIKey() error {
	return h.settings.SetSetting(h.ctx, settingSectorsAPIKey, "")
}
