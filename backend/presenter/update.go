package presenter

import (
	"context"
	"errors"
	"fmt"

	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/usecase"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const settingSkippedVersion = "skipped_version"

// updateAvailablePayload is the event payload emitted when a new version is detected on startup.
type updateAvailablePayload struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
	ReleaseURL     string `json:"releaseURL"`
}

// UpdateHandler handles update-related requests from the frontend.
type UpdateHandler struct {
	ctx        context.Context
	update     *usecase.UpdateService
	selfUpdate *usecase.SelfUpdateService
	settings   settings.Repository
	emitter    usecase.EventEmitter
}

// Bind wires the handler to its dependencies.
func (h *UpdateHandler) Bind(
	ctx context.Context,
	update *usecase.UpdateService,
	selfUpdate *usecase.SelfUpdateService,
	settings settings.Repository,
	emitter usecase.EventEmitter,
) {
	h.ctx = ctx
	h.update = update
	h.selfUpdate = selfUpdate
	h.settings = settings
	h.emitter = emitter
}

// CheckForUpdate checks for updates and returns the result for the frontend.
func (h *UpdateHandler) CheckForUpdate() (*UpdateCheckResponse, error) {
	result, err := h.update.Check(h.ctx)
	if err != nil {
		return nil, fmt.Errorf("check for update: %w", err)
	}
	return &UpdateCheckResponse{
		Available:      result.Available,
		CurrentVersion: result.CurrentVer,
		LatestVersion:  result.LatestVer,
		ReleaseURL:     result.ReleaseURL,
		ReleaseNotes:   result.ReleaseNotes,
	}, nil
}

// GetAppVersion returns the current application version.
func (h *UpdateHandler) GetAppVersion() string {
	return h.update.CurrentVersion()
}

// SkipVersion persists the given version so it won't trigger the startup notification again.
func (h *UpdateHandler) SkipVersion(version string) error {
	if version == "" {
		return errors.New("version must not be empty")
	}
	return h.settings.SetSetting(h.ctx, settingSkippedVersion, version)
}

// CheckForUpdateOnStartup checks for updates on app startup and emits an event if available.
// Skipped in dev builds to avoid noisy events during development.
func (h *UpdateHandler) CheckForUpdateOnStartup() {
	if h.update.CurrentVersion() == "dev" {
		return
	}

	result, err := h.update.Check(h.ctx)
	if err != nil {
		runtime.LogWarningf(h.ctx, "startup update check: %v", err)
		return
	}
	if !result.Available {
		return
	}

	skipped, _ := h.settings.GetSetting(h.ctx, settingSkippedVersion)
	if skipped == result.LatestVer {
		return
	}

	h.emitter.Emit(shared.EventUpdateAvailable, updateAvailablePayload{
		CurrentVersion: result.CurrentVer,
		LatestVersion:  result.LatestVer,
		ReleaseNotes:   result.ReleaseNotes,
		ReleaseURL:     result.ReleaseURL,
	})
}

// DownloadAndInstallUpdate starts the async self-update flow.
// Progress is reported via Wails events, not the return value.
func (h *UpdateHandler) DownloadAndInstallUpdate() error {
	go func() {
		if err := h.selfUpdate.PerformUpdate(h.ctx); err != nil {
			runtime.LogWarningf(h.ctx, "self-update: %v", err)
		}
	}()
	return nil
}

// CancelUpdate aborts an in-progress download.
func (h *UpdateHandler) CancelUpdate() {
	h.selfUpdate.Cancel()
}

// QuitForRestart gracefully shuts down the app so the user can relaunch.
func (h *UpdateHandler) QuitForRestart() {
	runtime.Quit(h.ctx)
}
