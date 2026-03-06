/**
 * Panen Alerts Store — Svelte 5 (Runes)
 *
 * Tracks fundamental change alerts via Wails events and backend calls.
 * Listens for alerts:updated event to refresh the active count.
 *
 * Usage in components:
 *   import { alerts } from "../stores/alerts.svelte";
 *
 *   <span>{alerts.activeCount}</span>
 */

import { EventsOn } from "../../../wailsjs/runtime/runtime";
import type { FundamentalAlertResponse } from "../types";

const browser = typeof window !== "undefined";

function createAlertsStore() {
  let activeCount = $state(0);
  let activeAlerts = $state<FundamentalAlertResponse[]>([]);
  let loading = $state(false);

  if (browser) {
    EventsOn("alerts:updated", (count: number) => {
      activeCount = count;
    });
  }

  async function loadActiveAlerts() {
    loading = true;
    try {
      const { GetActiveAlerts } = await import("../../../wailsjs/go/backend/App");
      const result = await GetActiveAlerts();
      activeAlerts = result ?? [];
      activeCount = activeAlerts.length;
    } catch {
      activeAlerts = [];
    } finally {
      loading = false;
    }
  }

  async function loadAlertsByTicker(ticker: string) {
    try {
      const { GetAlertsByTicker } = await import("../../../wailsjs/go/backend/App");
      const result = await GetAlertsByTicker(ticker);
      return result ?? [];
    } catch {
      return [];
    }
  }

  async function acknowledgeAlert(id: string) {
    try {
      const { AcknowledgeAlert } = await import("../../../wailsjs/go/backend/App");
      await AcknowledgeAlert(id);
      await loadActiveAlerts();
    } catch {
      // ignore
    }
  }

  async function loadCount() {
    try {
      const { GetAlertCount } = await import("../../../wailsjs/go/backend/App");
      activeCount = await GetAlertCount();
    } catch {
      // ignore
    }
  }

  return {
    get activeCount() {
      return activeCount;
    },
    get activeAlerts() {
      return activeAlerts;
    },
    get loading() {
      return loading;
    },
    loadActiveAlerts,
    loadAlertsByTicker,
    acknowledgeAlert,
    loadCount,
  };
}

export const alerts = createAlertsStore();
