<script lang="ts">
import { AlertCircle, CheckCircle2, Download, Loader2, X } from "lucide-svelte";
import {
  CancelUpdate,
  DownloadAndInstallUpdate,
  QuitForRestart,
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
</script>

<Modal
  open={updateStore.isActive || updateStore.state === "error"}
  aria-label={t("settings.updateDownloading")}
  onClose={dismiss}
  size="md"
>
  {#if updateStore.state === "downloading"}
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
  {/if}

  <!-- Close button for all states -->
  {#if updateStore.state !== "downloading" && updateStore.state !== "verifying" && updateStore.state !== "installing"}
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
