<script lang="ts">
import { onMount } from "svelte";
import {
  CheckForUpdate,
  GetAppVersion,
  GetRefreshSettings,
  OpenReleaseURL,
  TriggerRefresh,
  UpdateRefreshSettings,
} from "../../../wailsjs/go/backend/App";
import Alert from "../../lib/components/Alert.svelte";
import Button from "../../lib/components/Button.svelte";
import Select from "../../lib/components/Select.svelte";
import ThemeToggle from "../../lib/components/ThemeToggle.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import { sync } from "../../lib/stores/sync.svelte";
import { theme } from "../../lib/stores/theme.svelte";

let autoRefreshEnabled = $state(true);
let intervalMinutes = $state(720);
let lastRefreshedAt = $state("");
let loadError = $state<string | null>(null);
let saveError = $state<string | null>(null);

let appVersion = $state("");
let updateChecking = $state(false);
let updateResult = $state<{
  available: boolean;
  latestVersion: string;
  releaseURL: string;
} | null>(null);
let updateError = $state<string | null>(null);

onMount(async () => {
  try {
    const settings = await GetRefreshSettings();
    autoRefreshEnabled = settings.autoRefreshEnabled;
    intervalMinutes = settings.intervalMinutes;
    lastRefreshedAt = settings.lastRefreshedAt;
  } catch (e: unknown) {
    loadError = e instanceof Error ? e.message : String(e);
  }

  try {
    appVersion = await GetAppVersion();
  } catch {
    appVersion = "unknown";
  }
});

async function saveSettings() {
  saveError = null;
  try {
    await UpdateRefreshSettings(autoRefreshEnabled, intervalMinutes);
  } catch (e: unknown) {
    saveError = e instanceof Error ? e.message : String(e);
  }
}

async function triggerRefresh() {
  try {
    await TriggerRefresh();
  } catch {
    // error shown via sync store
  }
}

async function checkForUpdates() {
  updateChecking = true;
  updateResult = null;
  updateError = null;
  try {
    const result = await CheckForUpdate();
    updateResult = result;
  } catch (e: unknown) {
    updateError = e instanceof Error ? e.message : String(e);
  } finally {
    updateChecking = false;
  }
}

function openRelease(url: string) {
  OpenReleaseURL(url);
}
</script>

<div class="mx-auto max-w-lg px-4 py-8">
  <h2 class="mb-6 text-xl font-semibold text-text-primary">Settings</h2>

  {#if loadError}
    <Alert variant="negative" dismissible>Failed to load settings: {loadError}</Alert>
  {/if}

  {#if saveError}
    <Alert variant="negative" dismissible>Failed to save settings: {saveError}</Alert>
  {/if}

  <div class="space-y-6">
    <div>
      <label class="mb-1 block text-sm text-text-secondary" for="language">Language</label>
      <Select id="language" disabled class="opacity-60">
        <option>English</option>
        <option>Bahasa Indonesia</option>
      </Select>
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
          <Tooltip text="Automatically refresh stock data in the background at the configured interval">
            <span class="text-sm text-text-primary underline decoration-dotted cursor-help">Auto Refresh</span>
          </Tooltip>
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
          <Select
            id="refresh-interval"
            bind:value={intervalMinutes}
            onchange={() => saveSettings()}
            disabled={!autoRefreshEnabled}
          >
            <option value={180}>Every 3 hours</option>
            <option value={360}>Every 6 hours</option>
            <option value={720}>Every 12 hours</option>
            <option value={1440}>Every 24 hours</option>
          </Select>
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

    <div>
      <p class="mb-3 text-sm text-text-secondary">About</p>
      <div class="space-y-4 rounded-lg border border-border-default bg-bg-elevated p-4">
        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">Version</span>
          <span class="font-mono text-sm text-text-secondary">{appVersion}</span>
        </div>

        <Button
          variant="secondary"
          size="sm"
          loading={updateChecking}
          onclick={checkForUpdates}
        >
          Check for Updates
        </Button>

        {#if updateResult}
          {#if updateResult.available}
            <Alert variant="info">
              Panen {updateResult.latestVersion} is available.
              <button
                class="ml-1 font-medium underline underline-offset-2 hover:opacity-80"
                onclick={() => openRelease(updateResult!.releaseURL)}
              >
                View Release
              </button>
            </Alert>
          {:else}
            <Alert variant="positive">You're up to date.</Alert>
          {/if}
        {/if}

        {#if updateError}
          <Alert variant="negative" dismissible>
            Failed to check for updates: {updateError}
          </Alert>
        {/if}
      </div>
    </div>
  </div>

  <p class="mt-6 text-xs text-text-muted">Language selection coming in a future update</p>
</div>
