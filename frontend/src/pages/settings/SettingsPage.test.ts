import { render, screen, waitFor } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../i18n", () => ({
  t: (key: string, params?: Record<string, string | number>) => {
    const translations: Record<string, string> = {
      "nav.settings": "Settings",
      "settings.title": "Settings",
      "settings.theme": "Theme",
      "settings.language": "Language",
      "settings.english": "English",
      "settings.indonesian": "Bahasa Indonesia",
      "settings.dataRefresh": "Data Refresh",
      "settings.autoRefresh": "Auto Refresh",
      "settings.autoRefreshTooltip":
        "Automatically refresh stock data in the background at the configured interval",
      "settings.refreshInterval": "Refresh Interval",
      "settings.every3Hours": "Every 3 hours",
      "settings.every6Hours": "Every 6 hours",
      "settings.every12Hours": "Every 12 hours",
      "settings.every24Hours": "Every 24 hours",
      "settings.lastRefreshed": "Last refreshed:",
      "settings.refreshNow": "Refresh Now",
      "settings.syncing": "Syncing...",
      "settings.about": "About",
      "settings.version": "Version",
      "settings.checkForUpdates": "Check for Updates",
      "settings.upToDate": "You're up to date.",
      "settings.settingsSaved": "Settings saved",
      "settings.loadError": "Failed to load settings: {error}",
      "settings.saveError": "Failed to save settings: {error}",
      "settings.updateError": "Failed to check for updates: {error}",
      "settings.updateAvailable": "Panen {version} is available.",
      "settings.viewRelease": "View Release",
      "common.loading": "Loading...",
      "format.lastUpdated": "Last updated",
      "format.notSynced": "Not synced yet",
    };
    let value = translations[key] ?? key;
    if (params) {
      value = value.replace(/\{(\w+)\}/g, (_, name) => String(params[name] ?? `{${name}}`));
    }
    return value;
  },
  locale: {
    get current() {
      return "en";
    },
    set() {},
    toggle() {},
  },
}));

// Mock Wails runtime to avoid EventsOn error from SyncIndicator/sync store.
vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

// Mock theme store to avoid localStorage initialization issues in ThemeToggle.
vi.mock("../../lib/stores/theme.svelte", () => ({
  theme: {
    get preference() {
      return "light";
    },
    toggle() {},
    set() {},
  },
}));

const mockGetRefreshSettings = vi.fn();
const mockGetAppVersion = vi.fn();
const mockUpdateRefreshSettings = vi.fn();
const mockTriggerRefresh = vi.fn();
const mockCheckForUpdate = vi.fn();
const mockOpenReleaseURL = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetRefreshSettings: (...args: unknown[]) => mockGetRefreshSettings(...args),
  GetAppVersion: (...args: unknown[]) => mockGetAppVersion(...args),
  UpdateRefreshSettings: (...args: unknown[]) => mockUpdateRefreshSettings(...args),
  TriggerRefresh: (...args: unknown[]) => mockTriggerRefresh(...args),
  CheckForUpdate: (...args: unknown[]) => mockCheckForUpdate(...args),
  OpenReleaseURL: (...args: unknown[]) => mockOpenReleaseURL(...args),
}));

import SettingsPage from "./SettingsPage.svelte";

describe("SettingsPage", () => {
  beforeEach(() => {
    mockGetRefreshSettings.mockReset();
    mockGetAppVersion.mockReset();
    mockUpdateRefreshSettings.mockReset();
    mockTriggerRefresh.mockReset();
    mockCheckForUpdate.mockReset();
  });

  it("renders settings heading", async () => {
    mockGetRefreshSettings.mockResolvedValueOnce({
      autoRefreshEnabled: true,
      intervalMinutes: 720,
      lastRefreshedAt: "",
    });
    mockGetAppVersion.mockResolvedValueOnce("1.0.0");

    render(SettingsPage);

    expect(screen.getByText("Settings")).toBeInTheDocument();
  });

  it("loads and displays refresh settings", async () => {
    mockGetRefreshSettings.mockResolvedValueOnce({
      autoRefreshEnabled: true,
      intervalMinutes: 720,
      lastRefreshedAt: "2025-06-01T10:00:00Z",
    });
    mockGetAppVersion.mockResolvedValueOnce("1.2.0");

    render(SettingsPage);

    await waitFor(() => {
      expect(screen.getByText("1.2.0")).toBeInTheDocument();
    });
  });

  it("shows load error on failure", async () => {
    mockGetRefreshSettings.mockRejectedValueOnce(new Error("load failed"));
    mockGetAppVersion.mockResolvedValueOnce("1.0.0");

    render(SettingsPage);

    await waitFor(() => {
      expect(screen.getByText(/load failed/)).toBeInTheDocument();
    });
  });

  it("renders theme section", async () => {
    mockGetRefreshSettings.mockResolvedValueOnce({
      autoRefreshEnabled: true,
      intervalMinutes: 720,
      lastRefreshedAt: "",
    });
    mockGetAppVersion.mockResolvedValueOnce("1.0.0");

    render(SettingsPage);

    expect(screen.getByText("Theme")).toBeInTheDocument();
  });

  it("renders check for updates button", async () => {
    mockGetRefreshSettings.mockResolvedValueOnce({
      autoRefreshEnabled: true,
      intervalMinutes: 720,
      lastRefreshedAt: "",
    });
    mockGetAppVersion.mockResolvedValueOnce("1.0.0");

    render(SettingsPage);

    expect(screen.getByRole("button", { name: /Check for Updates/i })).toBeInTheDocument();
  });
});
