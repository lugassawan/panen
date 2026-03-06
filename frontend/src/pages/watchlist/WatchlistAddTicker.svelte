<script lang="ts">
import { Plus } from "lucide-svelte";
import { AddToWatchlist } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Button from "../../lib/components/Button.svelte";
import Input from "../../lib/components/Input.svelte";
import { toastStore } from "../../lib/stores/toast.svelte";

let {
  watchlistId,
  onAdded,
}: {
  watchlistId: string;
  onAdded: () => void;
} = $props();

let ticker = $state("");
let loading = $state(false);
let error = $state<string | null>(null);

async function submit(e: Event) {
  e.preventDefault();
  const trimmed = ticker.trim().toUpperCase();
  if (!trimmed) return;
  loading = true;
  error = null;
  try {
    await AddToWatchlist(watchlistId, trimmed);
    toastStore.add(`${trimmed} added to watchlist`, "success");
    ticker = "";
    onAdded();
  } catch (err: unknown) {
    error = err instanceof Error ? err.message : String(err);
  } finally {
    loading = false;
  }
}
</script>

<div class="border-b border-border-default px-6 py-3">
  <form onsubmit={submit} class="flex items-center gap-2">
    <div class="w-48">
      <Input
        bind:value={ticker}
        placeholder={t("watchlist.addTickerPlaceholder")}
        aria-label="Add ticker to watchlist"
        class="py-1.5 uppercase placeholder:normal-case placeholder:text-text-muted transition-fast"
        disabled={loading}
      />
    </div>
    <Button type="submit" size="sm" disabled={loading || !ticker.trim()} {loading}>
      <Plus size={14} strokeWidth={2} />
      {t("common.add")}
    </Button>
    {#if error}
      <p class="text-sm text-negative">{error}</p>
    {/if}
  </form>
</div>
