package presenter

import (
	"context"
	"fmt"
	"strings"

	"github.com/lugassawan/panen/backend/domain/settings"
	"github.com/lugassawan/panen/backend/usecase"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	settingSkippedVersion   = "skipped_version"
	allowedReleaseURLPrefix = "https://github.com/lugassawan/panen/releases/"
)

// UpdateHandler handles update-related requests from the frontend.
type UpdateHandler struct {
	ctx      context.Context
	update   *usecase.UpdateService
	settings settings.Repository
}

// NewUpdateHandler creates a new UpdateHandler.
func NewUpdateHandler(
	ctx context.Context,
	update *usecase.UpdateService,
	settings settings.Repository,
) *UpdateHandler {
	return &UpdateHandler{ctx: ctx, update: update, settings: settings}
}

// CheckForUpdate checks for updates and returns the result for the frontend.
func (h *UpdateHandler) CheckForUpdate() (*UpdateCheckResponse, error) {
	result, err := h.update.Check(h.ctx)
	if err != nil {
		return nil, err
	}
	return &UpdateCheckResponse{
		Available:      result.Available,
		CurrentVersion: result.CurrentVer,
		LatestVersion:  result.LatestVer,
		ReleaseURL:     result.ReleaseURL,
	}, nil
}

// GetAppVersion returns the current application version.
func (h *UpdateHandler) GetAppVersion() string {
	return h.update.CurrentVersion()
}

// OpenReleaseURL opens the given URL in the user's default browser.
// Only URLs under the project's GitHub releases path are allowed.
func (h *UpdateHandler) OpenReleaseURL(url string) {
	if !strings.HasPrefix(url, allowedReleaseURLPrefix) {
		runtime.LogWarningf(h.ctx, "blocked non-release URL: %s", url)
		return
	}
	runtime.BrowserOpenURL(h.ctx, url)
}

// CheckForUpdateOnStartup checks for updates on app startup and shows a native dialog if available.
// Skipped in dev builds to avoid noisy dialogs during development.
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

	msg := fmt.Sprintf(
		"A new version of Panen is available.\n\nCurrent: %s\nLatest: %s",
		result.CurrentVer, result.LatestVer,
	)

	selection, err := runtime.MessageDialog(h.ctx, runtime.MessageDialogOptions{
		Type:          runtime.InfoDialog,
		Title:         "Update Available",
		Message:       msg,
		Buttons:       []string{"View Release", "Skip This Version", "Dismiss"},
		DefaultButton: "View Release",
		CancelButton:  "Dismiss",
	})
	if err != nil {
		runtime.LogWarningf(h.ctx, "update dialog: %v", err)
		return
	}

	switch selection {
	case "View Release":
		runtime.BrowserOpenURL(h.ctx, result.ReleaseURL)
	case "Skip This Version":
		if err := h.settings.SetSetting(h.ctx, settingSkippedVersion, result.LatestVer); err != nil {
			runtime.LogWarningf(h.ctx, "save skipped version: %v", err)
		}
	}
}
