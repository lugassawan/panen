import { describe, expect, it, vi } from "vitest";

// Capture EventsOn callbacks
const eventHandlers: Record<string, (data: unknown) => void> = {};

// Mock Wails runtime
vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn((event: string, handler: (data: unknown) => void) => {
    eventHandlers[event] = handler;
  }),
}));

// Import after mock
const { updateStore } = await import("./update.svelte");

describe("updateStore", () => {
  it("starts in idle state", () => {
    expect(updateStore.state).toBe("idle");
    expect(updateStore.isActive).toBe(false);
    expect(updateStore.showNotification).toBe(false);
    expect(updateStore.progressPercent).toBe(0);
    expect(updateStore.version).toBe("");
    expect(updateStore.error).toBeNull();
    expect(updateStore.releaseNotes).toBe("");
    expect(updateStore.latestVersion).toBe("");
    expect(updateStore.currentVersion).toBe("");
  });

  it("reset returns to idle", () => {
    updateStore.reset();
    expect(updateStore.state).toBe("idle");
    expect(updateStore.downloadedBytes).toBe(0);
    expect(updateStore.totalBytes).toBe(0);
    expect(updateStore.showNotification).toBe(false);
  });

  it("progressPercent handles zero totalBytes", () => {
    expect(updateStore.progressPercent).toBe(0);
  });

  it("sets available state on update:available event", () => {
    const handler = eventHandlers["update:available"];
    expect(handler).toBeDefined();

    handler({
      currentVersion: "1.0.0",
      latestVersion: "1.1.0",
      releaseNotes: "- Cool feature\n- Bug fix",
      releaseURL: "https://github.com/lugassawan/panen/releases/tag/v1.1.0",
    });

    expect(updateStore.state).toBe("available");
    expect(updateStore.showNotification).toBe(true);
    expect(updateStore.isActive).toBe(false);
    expect(updateStore.currentVersion).toBe("1.0.0");
    expect(updateStore.latestVersion).toBe("1.1.0");
    expect(updateStore.releaseNotes).toBe("- Cool feature\n- Bug fix");
    expect(updateStore.releaseURL).toBe("https://github.com/lugassawan/panen/releases/tag/v1.1.0");

    // Clean up
    updateStore.reset();
  });
});
