import { render, screen } from "@testing-library/svelte";
import { userEvent } from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

const mockDownloadAndInstallUpdate = vi.fn();
const mockSkipVersion = vi.fn();

// Mock Wails bindings
vi.mock("../../../wailsjs/go/backend/App", () => ({
  CancelUpdate: vi.fn(),
  DownloadAndInstallUpdate: mockDownloadAndInstallUpdate,
  QuitForRestart: vi.fn(),
  SkipVersion: mockSkipVersion,
}));

// Capture EventsOn callbacks
const eventHandlers: Record<string, (data: unknown) => void> = {};

vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn((event: string, handler: (data: unknown) => void) => {
    eventHandlers[event] = handler;
  }),
}));

// Mock i18n
vi.mock("../../i18n", () => ({
  t: (key: string, params?: Record<string, string>) => {
    const translations: Record<string, string> = {
      "settings.updateDownloading": "Downloading update...",
      "settings.updateVerifying": "Verifying checksum...",
      "settings.updateInstalling": "Installing update...",
      "settings.updateReady": "Update installed. Restart to apply.",
      "settings.updateRestartNow": "Restart Now",
      "settings.updateLater": "Later",
      "settings.updateCancel": "Cancel",
      "settings.selfUpdateFailed": "Update failed",
      "settings.updateError": `Update failed: ${params?.error ?? ""}`,
      "settings.updateAvailableTitle": `Panen ${params?.version ?? ""} is available`,
      "settings.whatsChanged": "What's Changed",
      "settings.updateNow": "Update Now",
      "settings.skipThisVersion": "Skip This Version",
      "common.close": "Close",
    };
    return translations[key] ?? key;
  },
  locale: { current: "en" },
}));

vi.mock("../format", () => ({
  formatFileSize: (bytes: number) => `${bytes} B`,
}));

// Import component and store after mocks
const { default: UpdateDialog } = await import("./UpdateDialog.svelte");
const { updateStore } = await import("../stores/update.svelte");

describe("UpdateDialog", () => {
  it("renders nothing when idle", () => {
    const { container } = render(UpdateDialog);
    expect(container.querySelector("[role='dialog']")).toBeNull();
  });

  it("renders available state with release notes", () => {
    const handler = eventHandlers["update:available"];
    handler({
      currentVersion: "1.0.0",
      latestVersion: "1.1.0",
      releaseNotes: "- Add cool feature (#10)\n- Broken thing (#11)",
      releaseURL: "https://github.com/lugassawan/panen/releases/tag/v1.1.0",
    });

    render(UpdateDialog);

    expect(screen.getByText("Panen 1.1.0 is available")).toBeTruthy();
    expect(screen.getByText("What's Changed")).toBeTruthy();
    // Release notes are rendered in a <pre> tag
    const pre = document.querySelector("pre");
    expect(pre?.textContent).toContain("Add cool feature (#10)");
    expect(pre?.textContent).toContain("Broken thing (#11)");
    expect(screen.getByText("Update Now")).toBeTruthy();
    expect(screen.getByText("Skip This Version")).toBeTruthy();

    updateStore.reset();
  });

  it("calls DownloadAndInstallUpdate on Update Now click", async () => {
    const handler = eventHandlers["update:available"];
    handler({
      currentVersion: "1.0.0",
      latestVersion: "1.1.0",
      releaseNotes: "- Feature",
      releaseURL: "https://github.com/lugassawan/panen/releases/tag/v1.1.0",
    });

    render(UpdateDialog);

    const user = userEvent.setup();
    await user.click(screen.getByText("Update Now"));

    expect(mockDownloadAndInstallUpdate).toHaveBeenCalled();

    updateStore.reset();
  });

  it("calls SkipVersion on Skip This Version click", async () => {
    const handler = eventHandlers["update:available"];
    handler({
      currentVersion: "1.0.0",
      latestVersion: "1.1.0",
      releaseNotes: "- Feature",
      releaseURL: "https://github.com/lugassawan/panen/releases/tag/v1.1.0",
    });

    render(UpdateDialog);

    const user = userEvent.setup();
    await user.click(screen.getByText("Skip This Version"));

    expect(mockSkipVersion).toHaveBeenCalledWith("1.1.0");

    updateStore.reset();
  });
});
