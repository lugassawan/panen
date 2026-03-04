<script lang="ts">
import { onMount } from "svelte";
import {
  GetRefreshSettings,
  TriggerRefresh,
  UpdateRefreshSettings,
} from "../../wailsjs/go/backend/App";
import ThemeToggle from "../lib/components/ThemeToggle.svelte";
import { sync } from "../lib/stores/sync.svelte";
import { theme } from "../lib/stores/theme.svelte";

let autoRefreshEnabled = $state(true);
let intervalMinutes = $state(720);
let lastRefreshedAt = $state("");

onMount(async () => {
  try {
    const settings = await GetRefreshSettings();
    autoRefreshEnabled = settings.autoRefreshEnabled;
    intervalMinutes = settings.intervalMinutes;
    lastRefreshedAt = settings.lastRefreshedAt;
  } catch {
    // silently use defaults
  }
});

async function saveSettings() {
  try {
    await UpdateRefreshSettings(autoRefreshEnabled, intervalMinutes);
  } catch {
    // silently ignore save errors
  }
}

async function triggerRefresh() {
  try {
    await TriggerRefresh();
  } catch {
    // error shown via sync store
  }
}
</script>

<div class="mx-auto max-w-lg px-4 py-8">
  <h2 class="mb-6 text-xl font-semibold text-text-primary">Settings</h2>

  <div class="space-y-6">
    <div>
      <label class="mb-1 block text-sm text-text-secondary" for="language">Language</label>
      <select
        id="language"
        disabled
        class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary opacity-60"
      >
        <option>English</option>
        <option>Bahasa Indonesia</option>
      </select>
    </div>

    <div>
      <p class="mb-1 text-sm text-text-secondary">Theme</p>
      <div class="flex items-center gap-3">
        <ThemeToggle />
        <span class="text-sm text-text-tertiary capitalize">{theme.preference}</span>
      </div>
    </div>

    <div>
      <p class="mb-3 text-sm text-text-secondary">Data Refresh</p>
      <div class="space-y-4 rounded-lg border border-border-default bg-bg-elevated p-4">
        <!-- Auto Refresh Toggle -->
        <label class="flex items-center justify-between">
          <span class="text-sm text-text-primary">Auto Refresh</span>
          <input
            type="checkbox"
            bind:checked={autoRefreshEnabled}
            onchange={saveSettings}
            class="h-4 w-4 rounded border-border-default text-green-700 focus-ring"
          />
        </label>

        <!-- Interval Select -->
        <div>
          <label class="mb-1 block text-sm text-text-tertiary" for="refresh-interval">
            Refresh Interval
          </label>
          <select
            id="refresh-interval"
            bind:value={intervalMinutes}
            onchange={saveSettings}
            disabled={!autoRefreshEnabled}
            class="w-full rounded border border-border-default bg-bg-elevated px-3 py-2 text-sm text-text-primary disabled:opacity-60 focus-ring"
          >
            <option value={180}>Every 3 hours</option>
            <option value={360}>Every 6 hours</option>
            <option value={720}>Every 12 hours</option>
            <option value={1440}>Every 24 hours</option>
          </select>
        </div>

        <!-- Last Refreshed -->
        {#if lastRefreshedAt}
          <p class="text-xs text-text-tertiary">
            Last refreshed: <span class="font-mono">{lastRefreshedAt}</span>
          </p>
        {/if}

        <!-- Refresh Now Button -->
        <button
          onclick={triggerRefresh}
          disabled={sync.isSyncing}
          class="w-full rounded border border-green-700 px-3 py-2 text-sm font-medium text-green-700 transition-fast hover:bg-green-100 disabled:opacity-60 focus-ring dark:hover:bg-green-900/30"
        >
          {sync.isSyncing ? "Syncing..." : "Refresh Now"}
        </button>
      </div>
    </div>
  </div>

  <p class="mt-6 text-xs text-text-muted">Language selection coming in a future update</p>
</div>
