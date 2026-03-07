import { afterEach, describe, expect, it, vi } from "vitest";
import { EventRefreshProgress, EventRefreshStatus, EventRefreshSummary } from "../events";

// Capture event handlers registered by the store
const eventHandlers: Record<string, (...args: unknown[]) => void> = {};

vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn((event: string, callback: (...args: unknown[]) => void) => {
    eventHandlers[event] = callback;
  }),
}));

describe("sync store", () => {
  afterEach(() => {
    vi.resetModules();
    for (const key of Object.keys(eventHandlers)) {
      delete eventHandlers[key];
    }
  });

  async function loadSync() {
    const mod = await import("./sync.svelte");
    return mod.sync;
  }

  it("defaults to idle state with no progress", async () => {
    const sync = await loadSync();
    expect(sync.state).toBe("idle");
    expect(sync.isSyncing).toBe(false);
    expect(sync.lastRefresh).toBe("");
    expect(sync.currentTicker).toBeNull();
    expect(sync.progress).toBeNull();
    expect(sync.progressPercent).toBe(0);
    expect(sync.lastSummary).toBeNull();
    expect(sync.hasError).toBe(false);
    expect(sync.errorMessage).toBeNull();
  });

  it("updates state on refresh:status event", async () => {
    const sync = await loadSync();
    eventHandlers[EventRefreshStatus]({
      state: "syncing",
      lastRefresh: "2026-03-04T10:00:00Z",
    });
    expect(sync.state).toBe("syncing");
    expect(sync.isSyncing).toBe(true);
    expect(sync.lastRefresh).toBe("2026-03-04T10:00:00Z");
  });

  it("updates current ticker and progress on refresh:progress event", async () => {
    const sync = await loadSync();
    eventHandlers[EventRefreshProgress]({
      ticker: "BBCA",
      index: 2,
      total: 10,
      status: "success",
    });
    expect(sync.currentTicker).toBe("BBCA");
    expect(sync.progress).toEqual({
      ticker: "BBCA",
      index: 2,
      total: 10,
      status: "success",
    });
  });

  it("updates summary and clears progress on refresh:summary event", async () => {
    const sync = await loadSync();

    // First set some progress
    eventHandlers[EventRefreshProgress]({
      ticker: "BBRI",
      index: 4,
      total: 5,
      status: "success",
    });
    expect(sync.currentTicker).toBe("BBRI");

    // Then receive summary
    eventHandlers[EventRefreshSummary]({
      total: 5,
      fetched: 3,
      skipped: 1,
      failed: 1,
      duration: "2.5s",
    });
    expect(sync.lastSummary).toEqual({
      total: 5,
      fetched: 3,
      skipped: 1,
      failed: 1,
      duration: "2.5s",
    });
    expect(sync.progress).toBeNull();
    expect(sync.currentTicker).toBeNull();
  });

  it("computes progressPercent correctly", async () => {
    const sync = await loadSync();

    // index 0 of 4 → (0+1)/4 = 25%
    eventHandlers[EventRefreshProgress]({
      ticker: "TLKM",
      index: 0,
      total: 4,
      status: "success",
    });
    expect(sync.progressPercent).toBe(25);

    // index 1 of 4 → (1+1)/4 = 50%
    eventHandlers[EventRefreshProgress]({
      ticker: "ASII",
      index: 1,
      total: 4,
      status: "success",
    });
    expect(sync.progressPercent).toBe(50);

    // index 3 of 4 → (3+1)/4 = 100%
    eventHandlers[EventRefreshProgress]({
      ticker: "UNVR",
      index: 3,
      total: 4,
      status: "success",
    });
    expect(sync.progressPercent).toBe(100);
  });

  it("reflects syncing state via isSyncing", async () => {
    const sync = await loadSync();
    expect(sync.isSyncing).toBe(false);

    eventHandlers[EventRefreshStatus]({ state: "syncing", lastRefresh: "" });
    expect(sync.isSyncing).toBe(true);

    eventHandlers[EventRefreshStatus]({ state: "idle", lastRefresh: "" });
    expect(sync.isSyncing).toBe(false);
  });

  it("reflects error state via hasError and errorMessage", async () => {
    const sync = await loadSync();
    expect(sync.hasError).toBe(false);
    expect(sync.errorMessage).toBeNull();

    eventHandlers[EventRefreshStatus]({
      state: "error",
      lastRefresh: "",
      error: "network timeout",
    });
    expect(sync.hasError).toBe(true);
    expect(sync.errorMessage).toBe("network timeout");
  });
});
