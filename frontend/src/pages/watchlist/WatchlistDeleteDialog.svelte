<script lang="ts">
import { DeleteWatchlist } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Alert from "../../lib/components/Alert.svelte";
import ConfirmDialog from "../../lib/components/ConfirmDialog.svelte";
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
  <p>{t("common.confirmDeleteMessage", { name: watchlist.name })}</p>
  <p class="mt-1">{t("common.cannotUndo")}</p>
  {#if error}
    <div class="mt-3">
      <Alert variant="negative">{error}</Alert>
    </div>
  {/if}
</ConfirmDialog>
