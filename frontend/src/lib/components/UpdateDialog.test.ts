import { render } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";

// Mock Wails bindings
vi.mock("../../../wailsjs/go/backend/App", () => ({
  CancelUpdate: vi.fn(),
  DownloadAndInstallUpdate: vi.fn(),
  QuitForRestart: vi.fn(),
}));

vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
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
      "common.close": "Close",
    };
    return translations[key] ?? key;
  },
  locale: { current: "en" },
}));

vi.mock("../format", () => ({
  formatFileSize: (bytes: number) => `${bytes} B`,
}));

// Import component after mocks
const { default: UpdateDialog } = await import("./UpdateDialog.svelte");

describe("UpdateDialog", () => {
  it("renders nothing when idle", () => {
    const { container } = render(UpdateDialog);
    expect(container.querySelector("[role='dialog']")).toBeNull();
  });
});
