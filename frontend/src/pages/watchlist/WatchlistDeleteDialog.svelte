<script lang="ts">
import { DeleteWatchlist } from "../../../wailsjs/go/backend/App";
import ConfirmDialog from "../../components/ConfirmDialog.svelte";
import { t } from "../../i18n";
import type { WatchlistResponse } from "../../lib/types";

let {
  watchlist,
  onDeleted,
  onCancel,
}: {
  watchlist: WatchlistResponse;
  onDeleted: () => void;
  onCancel: () => void;
} = $props();

let loading = $state(false);
let error = $state<string | null>(null);

async function confirm() {
  loading = true;
  error = null;
  try {
    await DeleteWatchlist(watchlist.id);
    onDeleted();
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    loading = false;
  }
}
</script>

<ConfirmDialog
  title={t("watchlist.deleteTitle")}
  confirmLabel={t("common.delete")}
  confirmVariant="danger"
  {loading}
  onConfirm={confirm}
  {onCancel}
>
  <p>Are you sure you want to delete <strong>{watchlist.name}</strong>?</p>
  <p class="mt-1">{t("common.cannotUndo")}</p>
  {#if error}
    <div class="mt-3 rounded border border-negative/20 bg-negative-bg px-3 py-2 text-sm text-negative" role="alert">
      {error}
    </div>
  {/if}
</ConfirmDialog>
