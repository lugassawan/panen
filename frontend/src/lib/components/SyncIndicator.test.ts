import { cleanup, render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { mockTriggerRefresh, mockSync } = vi.hoisted(() => ({
  mockTriggerRefresh: vi.fn(() => Promise.resolve()),
  mockSync: {
    state: "idle" as string,
    isSyncing: false,
    lastRefresh: "",
    currentTicker: null as string | null,
    progress: null as { ticker: string; index: number; total: number } | null,
    progressPercent: 0,
    lastSummary: null,
    hasError: false,
    errorMessage: null as string | null,
  },
}));

vi.mock("../../../wailsjs/go/backend/App", () => ({
  TriggerRefresh: (...args: unknown[]) => mockTriggerRefresh(...args),
}));

vi.mock("../../../wailsjs/runtime/runtime", () => ({
  EventsOn: vi.fn(),
}));

vi.mock("../stores/sync.svelte", () => ({
  sync: mockSync,
}));

import SyncIndicator from "./SyncIndicator.svelte";

describe("SyncIndicator", () => {
  beforeEach(() => {
    mockSync.state = "idle";
    mockSync.isSyncing = false;
    mockSync.lastRefresh = "";
    mockSync.currentTicker = null;
    mockSync.progress = null;
    mockSync.progressPercent = 0;
    mockSync.lastSummary = null;
    mockSync.hasError = false;
    mockSync.errorMessage = null;
    mockTriggerRefresh.mockClear();
  });

  afterEach(() => {
    cleanup();
  });

  it("shows 'Not synced yet' when no lastRefresh", () => {
    render(SyncIndicator);
    expect(screen.getByText("Not synced yet")).toBeInTheDocument();
  });

  it("shows relative time when lastRefresh is set", () => {
    const fiveMinAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString();
    mockSync.lastRefresh = fiveMinAgo;
    render(SyncIndicator);
    expect(screen.getByText("5m ago")).toBeInTheDocument();
  });

  it("shows current ticker and progress when syncing", () => {
    mockSync.state = "syncing";
    mockSync.isSyncing = true;
    mockSync.currentTicker = "BBCA";
    mockSync.progress = { ticker: "BBCA", index: 2, total: 15 };
    mockSync.progressPercent = 20;
    render(SyncIndicator);
    expect(screen.getByText(/Syncing/)).toBeInTheDocument();
    expect(screen.getByText("BBCA", { exact: false })).toBeInTheDocument();
    expect(screen.getByText("(3/15)")).toBeInTheDocument();
  });

  it("shows error message and retry button on error", () => {
    mockSync.state = "error";
    mockSync.hasError = true;
    mockSync.errorMessage = "Network timeout";
    render(SyncIndicator);
    expect(screen.getByText("Network timeout")).toBeInTheDocument();
    expect(screen.getByText("Retry")).toBeInTheDocument();
  });

  it("calls TriggerRefresh when clicking retry", async () => {
    mockSync.state = "error";
    mockSync.hasError = true;
    mockSync.errorMessage = "Sync failed";
    const user = userEvent.setup();
    render(SyncIndicator);
    await user.click(screen.getByText("Retry"));
    expect(mockTriggerRefresh).toHaveBeenCalledOnce();
  });
});
