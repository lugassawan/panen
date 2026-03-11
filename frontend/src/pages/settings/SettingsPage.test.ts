import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
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
      "settings.backup": "Database Backup",
      "settings.lastBackup": "Last Backup",
      "settings.backupCount": "Backups",
      "settings.dbSize": "Database Size",
      "settings.totalBackupSize": "Backup Size",
      "settings.createBackup": "Create Backup",
      "settings.creatingBackup": "Creating...",
      "settings.backupCreated": "Backup created successfully",
      "settings.backupError": "Backup failed: {error}",
      "settings.noBackups": "No backups yet",
      "settings.backupTooltip":
        "Daily backups are created automatically on startup. Backups older than 7 days are removed.",
      "settings.debugAndLogs": "Debug & Logs",
      "settings.debugMode": "Debug Mode",
      "settings.debugModeTooltip":
        "Enable verbose logging for troubleshooting. Increases log file size.",
      "settings.debugEnabled": "Debug mode enabled",
      "settings.debugDisabled": "Debug mode disabled",
      "settings.logFiles": "Log Files",
      "settings.logSize": "Total Size",
      "settings.logDateRange": "Date Range",
      "settings.exportLogs": "Export Logs",
      "settings.exportingLogs": "Exporting...",
      "settings.logsExported": "Logs exported successfully",
      "settings.logsExportError": "Export failed: {error}",
      "settings.noLogs": "No logs yet",
      "settings.tab.general": "General",
      "settings.tab.data": "Data",
      "settings.tab.storage": "Storage",
      "settings.tab.system": "System",
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

vi.mock("../../lib/format", () => ({
  formatRelativeTime: (iso: string) => (iso ? "5m ago" : "Not synced yet"),
  formatFileSize: (bytes: number) => `${bytes} B`,
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
const mockGetBackupStatus = vi.fn();
const mockCreateManualBackup = vi.fn();
const mockIsDebugMode = vi.fn();
const mockSetDebugMode = vi.fn();
const mockExportLogs = vi.fn();
const mockGetLogStats = vi.fn();

vi.mock("../../../wailsjs/go/backend/App", () => ({
  GetRefreshSettings: (...args: unknown[]) => mockGetRefreshSettings(...args),
  GetAppVersion: (...args: unknown[]) => mockGetAppVersion(...args),
  UpdateRefreshSettings: (...args: unknown[]) => mockUpdateRefreshSettings(...args),
  TriggerRefresh: (...args: unknown[]) => mockTriggerRefresh(...args),
  CheckForUpdate: (...args: unknown[]) => mockCheckForUpdate(...args),
  GetBackupStatus: (...args: unknown[]) => mockGetBackupStatus(...args),
  CreateManualBackup: (...args: unknown[]) => mockCreateManualBackup(...args),
  IsDebugMode: (...args: unknown[]) => mockIsDebugMode(...args),
  SetDebugMode: (...args: unknown[]) => mockSetDebugMode(...args),
  ExportLogs: (...args: unknown[]) => mockExportLogs(...args),
  GetLogStats: (...args: unknown[]) => mockGetLogStats(...args),
}));

import SettingsPage from "./SettingsPage.svelte";

function setupMocks() {
  mockGetRefreshSettings.mockResolvedValue({
    autoRefreshEnabled: true,
    intervalMinutes: 720,
    lastRefreshedAt: "",
  });
  mockGetAppVersion.mockResolvedValue("1.0.0");
  mockGetBackupStatus.mockResolvedValue({
    lastBackupDate: "",
    backupCount: 0,
    totalSizeBytes: 0,
    dbSizeBytes: 0,
  });
  mockIsDebugMode.mockResolvedValue(false);
  mockGetLogStats.mockResolvedValue({
    fileCount: 0,
    totalBytes: 0,
    oldestDate: "",
    newestDate: "",
  });
}

describe("SettingsPage", () => {
  beforeEach(() => {
    mockGetRefreshSettings.mockReset();
    mockGetAppVersion.mockReset();
    mockUpdateRefreshSettings.mockReset();
    mockTriggerRefresh.mockReset();
    mockCheckForUpdate.mockReset();
    mockGetBackupStatus.mockReset();
    mockCreateManualBackup.mockReset();
    mockIsDebugMode.mockReset();
    mockSetDebugMode.mockReset();
    mockExportLogs.mockReset();
    mockGetLogStats.mockReset();
    setupMocks();
  });

  it("renders settings heading and tab bar", async () => {
    render(SettingsPage);

    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "General" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Data" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Storage" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "System" })).toBeInTheDocument();
  });

  it("shows General tab by default with theme and language", async () => {
    render(SettingsPage);

    expect(screen.getByRole("tab", { name: "General" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByText("Theme")).toBeInTheDocument();
    expect(screen.getByText("Language")).toBeInTheDocument();
  });

  it("loads and displays refresh settings on Data tab", async () => {
    const user = userEvent.setup();
    mockGetRefreshSettings.mockResolvedValue({
      autoRefreshEnabled: true,
      intervalMinutes: 720,
      lastRefreshedAt: "2025-06-01T10:00:00Z",
    });
    mockGetAppVersion.mockResolvedValue("1.2.0");

    render(SettingsPage);
    await user.click(screen.getByRole("tab", { name: "Data" }));

    await waitFor(() => {
      expect(screen.getByText("5m ago")).toBeInTheDocument();
    });
  });

  it("shows load error on failure", async () => {
    mockGetRefreshSettings.mockRejectedValue(new Error("load failed"));

    render(SettingsPage);

    await waitFor(() => {
      expect(screen.getByText(/load failed/)).toBeInTheDocument();
    });
  });

  it("renders check for updates button on System tab", async () => {
    const user = userEvent.setup();
    render(SettingsPage);
    await user.click(screen.getByRole("tab", { name: "System" }));

    expect(screen.getByRole("button", { name: /Check for Updates/i })).toBeInTheDocument();
  });

  it("renders backup section on Storage tab", async () => {
    const user = userEvent.setup();
    mockGetBackupStatus.mockResolvedValue({
      lastBackupDate: "2026-03-07T10:00:00Z",
      backupCount: 3,
      totalSizeBytes: 2048,
      dbSizeBytes: 4096,
    });

    render(SettingsPage);
    await user.click(screen.getByRole("tab", { name: "Storage" }));

    await waitFor(() => {
      expect(screen.getByText("Database Backup")).toBeInTheDocument();
      expect(screen.getByText("3")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Create Backup/i })).toBeInTheDocument();
    });
  });

  it("renders debug and logs section on System tab", async () => {
    const user = userEvent.setup();
    render(SettingsPage);
    await user.click(screen.getByRole("tab", { name: "System" }));

    await waitFor(() => {
      expect(screen.getByText("Debug & Logs")).toBeInTheDocument();
      expect(screen.getByText("Debug Mode")).toBeInTheDocument();
      expect(screen.getByText("Log Files")).toBeInTheDocument();
      expect(screen.getByText("Total Size")).toBeInTheDocument();
    });
  });

  it("renders log stats with file count and date range on System tab", async () => {
    const user = userEvent.setup();
    mockGetLogStats.mockResolvedValue({
      fileCount: 5,
      totalBytes: 10240,
      oldestDate: "2026-03-01",
      newestDate: "2026-03-07",
    });

    render(SettingsPage);
    await user.click(screen.getByRole("tab", { name: "System" }));

    await waitFor(() => {
      expect(screen.getByText("5")).toBeInTheDocument();
      expect(screen.getByText("10240 B")).toBeInTheDocument();
      expect(screen.getByText("2026-03-01 — 2026-03-07")).toBeInTheDocument();
    });
  });

  it("renders export logs button on System tab", async () => {
    const user = userEvent.setup();
    mockGetLogStats.mockResolvedValue({
      fileCount: 2,
      totalBytes: 512,
      oldestDate: "2026-03-06",
      newestDate: "2026-03-07",
    });

    render(SettingsPage);
    await user.click(screen.getByRole("tab", { name: "System" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Export Logs/i })).toBeInTheDocument();
    });
  });

  it("calls SetDebugMode when debug checkbox is toggled on System tab", async () => {
    const user = userEvent.setup();
    mockSetDebugMode.mockResolvedValue(undefined);

    render(SettingsPage);
    await user.click(screen.getByRole("tab", { name: "System" }));

    await waitFor(() => {
      expect(screen.getByText("Debug Mode")).toBeInTheDocument();
    });

    const checkboxes = screen.getAllByRole("checkbox");
    const debugCheckbox = checkboxes.find(
      (cb) => !cb.closest("label")?.textContent?.includes("Auto Refresh"),
    );
    if (debugCheckbox) {
      debugCheckbox.click();
      await waitFor(() => {
        expect(mockSetDebugMode).toHaveBeenCalledWith(true);
      });
    }
  });
});
