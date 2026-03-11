<script lang="ts">
import { AlertCircle, ArrowRight, CheckCircle2, Download, Info, Loader2, X } from "lucide-svelte";
import {
  CancelUpdate,
  DownloadAndInstallUpdate,
  QuitForRestart,
  SkipVersion,
} from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import { formatFileSize } from "../format";
import { updateStore } from "../stores/update.svelte";

import Button from "./Button.svelte";
import Modal from "./Modal.svelte";

function dismiss() {
  updateStore.reset();
}

function cancel() {
  CancelUpdate();
  updateStore.reset();
}

function restart() {
  QuitForRestart();
}

function startUpdate() {
  DownloadAndInstallUpdate();
}

async function skipVersion() {
  const version = updateStore.latestVersion;
  if (version) {
    await SkipVersion(version);
  }
  dismiss();
}
</script>

<Modal
  open={updateStore.showNotification || updateStore.isActive || updateStore.state === "error"}
  aria-label={updateStore.showNotification ? t("settings.updateAvailableTitle") : t("settings.updateDownloading")}
  onClose={dismiss}
  size="md"
>
  {#if updateStore.state === "available"}
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <Info size={20} class="text-info shrink-0" />
        <h3 class="text-lg font-semibold text-text-primary">
          {t("settings.updateAvailableTitle", { version: updateStore.latestVersion })}
        </h3>
      </div>

      <div class="flex items-center gap-2 text-sm text-text-secondary">
        <span class="font-mono">{updateStore.currentVersion}</span>
        <ArrowRight size={14} class="text-text-tertiary" />
        <span class="font-mono font-semibold text-text-primary">{updateStore.latestVersion}</span>
      </div>

      {#if updateStore.releaseNotes}
        <div>
          <p class="mb-2 text-xs font-medium text-text-secondary uppercase tracking-wide">
            {t("settings.whatsChanged")}
          </p>
          <div class="max-h-60 overflow-y-auto rounded border border-border-default bg-bg-secondary p-3">
            <pre class="whitespace-pre-wrap font-mono text-xs text-text-primary leading-relaxed">{updateStore.releaseNotes}</pre>
          </div>
        </div>
      {/if}

      <div class="flex justify-end gap-2">
        <Button variant="ghost" size="sm" onclick={skipVersion}>
          {t("settings.skipThisVersion")}
        </Button>
        <Button variant="primary" size="sm" onclick={startUpdate}>
          {t("settings.updateNow")}
        </Button>
      </div>
    </div>

    <button
      type="button"
      class="absolute top-4 right-4 text-text-tertiary hover:text-text-primary transition-fast focus-ring rounded"
      onclick={dismiss}
      aria-label={t("common.close")}
    >
      <X size={16} />
    </button>

  {:else if updateStore.state === "downloading"}
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <Download size={20} class="text-info shrink-0" />
        <h3 class="text-lg font-semibold text-text-primary">
          {t("settings.updateDownloading")}
        </h3>
      </div>

      <div class="space-y-2">
        <div class="h-2 w-full overflow-hidden rounded-full bg-bg-secondary">
          <div
            class="h-full rounded-full bg-green-600 transition-all duration-300"
            style="width: {updateStore.progressPercent}%"
          ></div>
        </div>

        <div class="flex justify-between text-xs text-text-tertiary">
          <span class="font-mono">
            {formatFileSize(updateStore.downloadedBytes)} / {formatFileSize(updateStore.totalBytes)}
          </span>
          <span class="font-mono">{updateStore.progressPercent}%</span>
        </div>
      </div>

      <div class="flex justify-end">
        <Button variant="secondary" size="sm" onclick={cancel}>
          {t("settings.updateCancel")}
        </Button>
      </div>
    </div>

  {:else if updateStore.state === "verifying"}
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <Loader2 size={20} class="text-info shrink-0 animate-spin" />
        <h3 class="text-lg font-semibold text-text-primary">
          {t("settings.updateVerifying")}
        </h3>
      </div>
    </div>

  {:else if updateStore.state === "installing"}
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <Loader2 size={20} class="text-info shrink-0 animate-spin" />
        <h3 class="text-lg font-semibold text-text-primary">
          {t("settings.updateInstalling")}
        </h3>
      </div>
    </div>

  {:else if updateStore.state === "ready"}
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <CheckCircle2 size={20} class="text-profit shrink-0" />
        <h3 class="text-lg font-semibold text-text-primary">
          {t("settings.updateReady")}
        </h3>
      </div>

      <div class="flex justify-end gap-2">
        <Button variant="secondary" size="sm" onclick={dismiss}>
          {t("settings.updateLater")}
        </Button>
        <Button variant="primary" size="sm" onclick={restart}>
          {t("settings.updateRestartNow")}
        </Button>
      </div>
    </div>

    <button
      type="button"
      class="absolute top-4 right-4 text-text-tertiary hover:text-text-primary transition-fast focus-ring rounded"
      onclick={dismiss}
      aria-label={t("common.close")}
    >
      <X size={16} />
    </button>

  {:else if updateStore.state === "error"}
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <AlertCircle size={20} class="text-loss shrink-0" />
        <h3 class="text-lg font-semibold text-text-primary">
          {t("settings.selfUpdateFailed")}
        </h3>
      </div>

      {#if updateStore.error}
        <p class="text-sm text-text-secondary break-words">
          {updateStore.error}
        </p>
      {/if}

      <div class="flex justify-end">
        <Button variant="secondary" size="sm" onclick={dismiss}>
          {t("common.close")}
        </Button>
      </div>
    </div>

    <button
      type="button"
      class="absolute top-4 right-4 text-text-tertiary hover:text-text-primary transition-fast focus-ring rounded"
      onclick={dismiss}
      aria-label={t("common.close")}
    >
      <X size={16} />
    </button>
  {/if}
</Modal>
