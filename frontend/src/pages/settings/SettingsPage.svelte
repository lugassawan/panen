<script lang="ts">
import { onMount } from "svelte";
import {
  CheckForUpdate,
  CreateManualBackup,
  DownloadAndInstallUpdate,
  ExportLogs,
  GetAppVersion,
  GetBackupStatus,
  GetLogStats,
  GetRefreshSettings,
  IsDebugMode,
  OpenReleaseURL,
  SetDebugMode,
  TriggerRefresh,
  UpdateRefreshSettings,
} from "../../../wailsjs/go/backend/App";
import type { Locale } from "../../i18n";
import { locale, t } from "../../i18n";
import Alert from "../../lib/components/Alert.svelte";
import Button from "../../lib/components/Button.svelte";
import Select from "../../lib/components/Select.svelte";
import ThemeToggle from "../../lib/components/ThemeToggle.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import UpdateDialog from "../../lib/components/UpdateDialog.svelte";
import { formatFileSize, formatRelativeTime } from "../../lib/format";
import { sync } from "../../lib/stores/sync.svelte";
import { theme } from "../../lib/stores/theme.svelte";
import { toastStore } from "../../lib/stores/toast.svelte";
import { updateStore } from "../../lib/stores/update.svelte";

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

let backupStatus = $state<{
  lastBackupDate: string;
  backupCount: number;
  totalSizeBytes: number;
  dbSizeBytes: number;
} | null>(null);
let backupCreating = $state(false);

let debugMode = $state(false);
let logStats = $state<{
  fileCount: number;
  totalBytes: number;
  oldestDate: string;
  newestDate: string;
} | null>(null);
let exportingLogs = $state(false);

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

  try {
    backupStatus = await GetBackupStatus();
  } catch {
    // non-critical
  }

  try {
    debugMode = await IsDebugMode();
  } catch {
    // non-critical
  }

  try {
    logStats = await GetLogStats();
  } catch {
    // non-critical
  }
});

async function saveSettings() {
  saveError = null;
  try {
    await UpdateRefreshSettings(autoRefreshEnabled, intervalMinutes);
    toastStore.add(t("settings.settingsSaved"), "success");
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

async function toggleDebugMode() {
  const newValue = !debugMode;
  try {
    await SetDebugMode(newValue);
    debugMode = newValue;
    toastStore.add(t(newValue ? "settings.debugEnabled" : "settings.debugDisabled"), "success");
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    toastStore.add(msg, "error");
  }
}

async function exportLogs() {
  exportingLogs = true;
  try {
    const path = await ExportLogs();
    if (path) {
      toastStore.add(t("settings.logsExported"), "success");
    }
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    toastStore.add(t("settings.logsExportError", { error: msg }), "error");
  } finally {
    exportingLogs = false;
  }
}

async function createBackup() {
  backupCreating = true;
  try {
    await CreateManualBackup();
    toastStore.add(t("settings.backupCreated"), "success");
    backupStatus = await GetBackupStatus();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    toastStore.add(t("settings.backupError", { error: msg }), "error");
  } finally {
    backupCreating = false;
  }
}
</script>

<div class="mx-auto max-w-lg px-4 py-8">
  <h1 class="mb-6 text-2xl font-display font-bold text-text-primary">{t("settings.title")}</h1>

  {#if loadError}
    <Alert variant="negative" dismissible>{t("settings.loadError", { error: loadError })}</Alert>
  {/if}

  {#if saveError}
    <Alert variant="negative" dismissible>{t("settings.saveError", { error: saveError })}</Alert>
  {/if}

  <div class="space-y-6">
    <div>
      <label class="mb-1 block text-sm text-text-secondary" for="language">{t("settings.language")}</label>
      <Select id="language" value={locale.current} onchange={(e) => locale.set(e.currentTarget.value as Locale)}>
        <option value="en">{t("settings.english")}</option>
        <option value="id">{t("settings.indonesian")}</option>
      </Select>
    </div>

    <div>
      <p class="mb-1 text-sm text-text-secondary">{t("settings.theme")}</p>
      <div class="flex items-center gap-3">
        <ThemeToggle />
        <span class="text-sm text-text-tertiary capitalize">{theme.preference}</span>
      </div>
    </div>

    <div>
      <p class="mb-3 text-sm text-text-secondary">{t("settings.dataRefresh")}</p>
      <div class="space-y-4 rounded-lg border border-border-default bg-bg-elevated p-4">
        <label class="flex items-center justify-between">
          <Tooltip text={t("settings.autoRefreshTooltip")}>
            <span class="text-sm text-text-primary underline decoration-dotted cursor-help">{t("settings.autoRefresh")}</span>
          </Tooltip>
          <input
            type="checkbox"
            bind:checked={autoRefreshEnabled}
            onchange={saveSettings}
            class="h-4 w-4 rounded border-border-default text-green-700 focus-ring"
          />
        </label>

        <div>
          <label class="mb-1 block text-sm text-text-tertiary" for="refresh-interval">
            {t("settings.refreshInterval")}
          </label>
          <Select
            id="refresh-interval"
            bind:value={intervalMinutes}
            onchange={() => saveSettings()}
            disabled={!autoRefreshEnabled}
          >
            <option value={180}>{t("settings.every3Hours")}</option>
            <option value={360}>{t("settings.every6Hours")}</option>
            <option value={720}>{t("settings.every12Hours")}</option>
            <option value={1440}>{t("settings.every24Hours")}</option>
          </Select>
        </div>

        {#if lastRefreshedAt}
          <p class="text-xs text-text-tertiary">
            {t("settings.lastRefreshed")} <span class="font-mono">{formatRelativeTime(lastRefreshedAt)}</span>
          </p>
        {/if}

        <button
          onclick={triggerRefresh}
          disabled={sync.isSyncing}
          class="w-full rounded border border-green-700 px-3 py-2 text-sm font-medium text-green-700 transition-fast hover:bg-green-100 disabled:opacity-60 focus-ring dark:hover:bg-green-900/30"
        >
          {sync.isSyncing ? t("settings.syncing") : t("settings.refreshNow")}
        </button>
      </div>
    </div>

    <div>
      <p class="mb-3 text-sm text-text-secondary">
        <Tooltip text={t("settings.backupTooltip")}>
          <span class="underline decoration-dotted cursor-help">{t("settings.backup")}</span>
        </Tooltip>
      </p>
      <div class="space-y-4 rounded-lg border border-border-default bg-bg-elevated p-4">
        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">{t("settings.lastBackup")}</span>
          <span class="font-mono text-sm text-text-secondary">
            {#if backupStatus?.lastBackupDate}
              {formatRelativeTime(backupStatus.lastBackupDate)}
            {:else}
              {t("settings.noBackups")}
            {/if}
          </span>
        </div>

        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">{t("settings.backupCount")}</span>
          <span class="font-mono text-sm text-text-secondary">{backupStatus?.backupCount ?? 0}</span>
        </div>

        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">{t("settings.dbSize")}</span>
          <span class="font-mono text-sm text-text-secondary">{formatFileSize(backupStatus?.dbSizeBytes ?? 0)}</span>
        </div>

        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">{t("settings.totalBackupSize")}</span>
          <span class="font-mono text-sm text-text-secondary">{formatFileSize(backupStatus?.totalSizeBytes ?? 0)}</span>
        </div>

        <button
          onclick={createBackup}
          disabled={backupCreating}
          class="w-full rounded border border-green-700 px-3 py-2 text-sm font-medium text-green-700 transition-fast hover:bg-green-100 disabled:opacity-60 focus-ring dark:hover:bg-green-900/30"
        >
          {backupCreating ? t("settings.creatingBackup") : t("settings.createBackup")}
        </button>
      </div>
    </div>

    <div>
      <p class="mb-3 text-sm text-text-secondary">{t("settings.debugAndLogs")}</p>
      <div class="space-y-4 rounded-lg border border-border-default bg-bg-elevated p-4">
        <label class="flex items-center justify-between">
          <Tooltip text={t("settings.debugModeTooltip")}>
            <span class="text-sm text-text-primary underline decoration-dotted cursor-help">{t("settings.debugMode")}</span>
          </Tooltip>
          <input
            type="checkbox"
            checked={debugMode}
            onchange={toggleDebugMode}
            class="h-4 w-4 rounded border-border-default text-green-700 focus-ring"
          />
        </label>

        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">{t("settings.logFiles")}</span>
          <span class="font-mono text-sm text-text-secondary">
            {logStats?.fileCount ?? 0}
          </span>
        </div>

        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">{t("settings.logSize")}</span>
          <span class="font-mono text-sm text-text-secondary">
            {formatFileSize(logStats?.totalBytes ?? 0)}
          </span>
        </div>

        {#if logStats && logStats.oldestDate && logStats.newestDate}
          <div class="flex items-center justify-between">
            <span class="text-sm text-text-primary">{t("settings.logDateRange")}</span>
            <span class="font-mono text-sm text-text-secondary">
              {logStats.oldestDate} — {logStats.newestDate}
            </span>
          </div>
        {/if}

        <button
          onclick={exportLogs}
          disabled={exportingLogs || (logStats?.fileCount ?? 0) === 0}
          class="w-full rounded border border-green-700 px-3 py-2 text-sm font-medium text-green-700 transition-fast hover:bg-green-100 disabled:opacity-60 focus-ring dark:hover:bg-green-900/30"
        >
          {exportingLogs ? t("settings.exportingLogs") : t("settings.exportLogs")}
        </button>
      </div>
    </div>

    <div>
      <p class="mb-3 text-sm text-text-secondary">{t("settings.about")}</p>
      <div class="space-y-4 rounded-lg border border-border-default bg-bg-elevated p-4">
        <div class="flex items-center justify-between">
          <span class="text-sm text-text-primary">{t("settings.version")}</span>
          <span class="font-mono text-sm text-text-secondary">{appVersion}</span>
        </div>

        <Button
          variant="secondary"
          size="sm"
          loading={updateChecking}
          onclick={checkForUpdates}
        >
          {t("settings.checkForUpdates")}
        </Button>

        {#if updateResult}
          {#if updateResult.available}
            <Alert variant="info">
              {t("settings.updateAvailable", { version: updateResult.latestVersion })}
              <button
                class="ml-1 font-medium underline underline-offset-2 hover:opacity-80"
                onclick={() => openRelease(updateResult!.releaseURL)}
              >
                {t("settings.viewRelease")}
              </button>
            </Alert>
            <Button
              variant="primary"
              size="sm"
              onclick={() => DownloadAndInstallUpdate()}
              disabled={updateStore.isActive}
            >
              {t("settings.downloadAndInstall")}
            </Button>
          {:else}
            <Alert variant="positive">{t("settings.upToDate")}</Alert>
          {/if}
        {/if}

        {#if updateError}
          <Alert variant="negative" dismissible>
            {t("settings.updateError", { error: updateError })}
          </Alert>
        {/if}
      </div>
    </div>
  </div>
</div>

<UpdateDialog />
