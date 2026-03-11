/**
 * Panen Update Store — Svelte 5 (Runes)
 *
 * Tracks self-update progress via Wails events and startup update notifications.
 *
 * Usage in components:
 *   import { updateStore } from "../stores/update.svelte";
 *
 *   <span>{updateStore.state}</span>
 *   <span>{updateStore.progressPercent}</span>
 */

import { EventsOn } from "../../../wailsjs/runtime/runtime";
import { EventUpdateAvailable, EventUpdateProgress } from "../events";

export type UpdateState =
  | "idle"
  | "available"
  | "downloading"
  | "verifying"
  | "installing"
  | "ready"
  | "error"
  | "cancelled";

export interface UpdateProgress {
  state: UpdateState;
  downloadedBytes: number;
  totalBytes: number;
  version: string;
  error?: string;
}

export interface UpdateAvailablePayload {
  currentVersion: string;
  latestVersion: string;
  releaseNotes: string;
  releaseURL: string;
}

const browser = typeof window !== "undefined";

function createUpdateStore() {
  let progress = $state<UpdateProgress>({
    state: "idle",
    downloadedBytes: 0,
    totalBytes: 0,
    version: "",
  });

  let notification = $state<UpdateAvailablePayload | null>(null);

  if (browser) {
    EventsOn(EventUpdateProgress, (data: UpdateProgress) => {
      progress = data;
    });

    EventsOn(EventUpdateAvailable, (data: UpdateAvailablePayload) => {
      notification = data;
      progress = {
        state: "available",
        downloadedBytes: 0,
        totalBytes: 0,
        version: data.latestVersion,
      };
    });
  }

  return {
    get state() {
      return progress.state;
    },
    get downloadedBytes() {
      return progress.downloadedBytes;
    },
    get totalBytes() {
      return progress.totalBytes;
    },
    get progressPercent() {
      if (progress.totalBytes <= 0) return 0;
      return Math.round((progress.downloadedBytes / progress.totalBytes) * 100);
    },
    get version() {
      return progress.version;
    },
    get error() {
      return progress.error ?? null;
    },
    get isActive() {
      return (
        progress.state === "downloading" ||
        progress.state === "verifying" ||
        progress.state === "installing" ||
        progress.state === "ready"
      );
    },
    get showNotification() {
      return progress.state === "available";
    },
    get releaseNotes() {
      return notification?.releaseNotes ?? "";
    },
    get latestVersion() {
      return notification?.latestVersion ?? "";
    },
    get currentVersion() {
      return notification?.currentVersion ?? "";
    },
    get releaseURL() {
      return notification?.releaseURL ?? "";
    },
    reset() {
      notification = null;
      progress = {
        state: "idle",
        downloadedBytes: 0,
        totalBytes: 0,
        version: "",
      };
    },
  };
}

export const updateStore = createUpdateStore();
