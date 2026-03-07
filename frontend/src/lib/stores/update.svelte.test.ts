import { describe, expect, it, vi } from "vitest";

// Mock Wails runtime
vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

// Import after mock
const { updateStore } = await import("./update.svelte");

describe("updateStore", () => {
  it("starts in idle state", () => {
    expect(updateStore.state).toBe("idle");
    expect(updateStore.isActive).toBe(false);
    expect(updateStore.progressPercent).toBe(0);
    expect(updateStore.version).toBe("");
    expect(updateStore.error).toBeNull();
  });

  it("reset returns to idle", () => {
    updateStore.reset();
    expect(updateStore.state).toBe("idle");
    expect(updateStore.downloadedBytes).toBe(0);
    expect(updateStore.totalBytes).toBe(0);
  });

  it("progressPercent handles zero totalBytes", () => {
    expect(updateStore.progressPercent).toBe(0);
  });
});
