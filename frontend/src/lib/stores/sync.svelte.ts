/**
 * Panen Sync Store — Svelte 5 (Runes)
 *
 * Tracks background refresh status via Wails events.
 * Listens for refresh:status, refresh:progress, and refresh:summary events.
 *
 * Usage in components:
 *   import { sync } from "../stores/sync.svelte";
 *
 *   <span>{sync.state}</span>           // 'idle' | 'syncing' | 'error'
 *   <span>{sync.currentTicker}</span>   // ticker being processed or null
 *   <span>{sync.progressPercent}</span> // 0–100
 */

import { EventsOn } from "../../../wailsjs/runtime/runtime";
import type { RefreshProgress, RefreshStatus, RefreshSummary } from "../types";

const browser = typeof window !== "undefined";

function createSyncStore() {
  let status = $state<RefreshStatus>({ state: "idle", lastRefresh: "" });
  let currentProgress = $state<RefreshProgress | null>(null);
  let lastSummary = $state<RefreshSummary | null>(null);

  if (browser) {
    EventsOn("refresh:status", (data: RefreshStatus) => {
      status = data;
    });
    EventsOn("refresh:progress", (data: RefreshProgress) => {
      currentProgress = data;
    });
    EventsOn("refresh:summary", (data: RefreshSummary) => {
      lastSummary = data;
      currentProgress = null;
    });
  }

  return {
    get state() {
      return status.state;
    },
    get isSyncing() {
      return status.state === "syncing";
    },
    get lastRefresh() {
      return status.lastRefresh;
    },
    get currentTicker() {
      return currentProgress?.ticker ?? null;
    },
    get progress() {
      return currentProgress;
    },
    get progressPercent() {
      if (!currentProgress) return 0;
      return Math.round(((currentProgress.index + 1) / currentProgress.total) * 100);
    },
    get lastSummary() {
      return lastSummary;
    },
    get hasError() {
      return status.state === "error";
    },
    get errorMessage() {
      return status.error ?? null;
    },
  };
}

export const sync = createSyncStore();
