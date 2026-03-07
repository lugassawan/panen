<script lang="ts">
import { ArrowLeftRight, LoaderCircle, Minus, Plus } from "lucide-svelte";
import { GetWatchlistItems, ListWatchlists, LookupStock } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import EmptyState from "../../lib/components/EmptyState.svelte";
import Input from "../../lib/components/Input.svelte";
import Select from "../../lib/components/Select.svelte";
import type {
  RiskProfile,
  StockValuationResponse,
  WatchlistItemResponse,
  WatchlistResponse,
} from "../../lib/types";
import ComparisonTable from "./ComparisonTable.svelte";

interface TickerSlot {
  code: string;
  riskProfile: RiskProfile;
  result: StockValuationResponse | null;
  loading: boolean;
  error: string | null;
}

function createSlot(): TickerSlot {
  return { code: "", riskProfile: "MODERATE", result: null, loading: false, error: null };
}

let slots = $state<TickerSlot[]>([createSlot(), createSlot()]);
let comparing = $state(false);
let hasCompared = $state(false);

let watchlists = $state<WatchlistResponse[]>([]);
let watchlistItems = $state<Map<string, WatchlistItemResponse[]>>(new Map());

$effect(() => {
  ListWatchlists().then((wl) => {
    watchlists = wl ?? [];
  });
});

async function loadWatchlistItems(watchlistId: string): Promise<WatchlistItemResponse[]> {
  const cached = watchlistItems.get(watchlistId);
  if (cached) return cached;
  const items = (await GetWatchlistItems(watchlistId)) ?? [];
  watchlistItems.set(watchlistId, items);
  return items;
}

function addSlot() {
  if (slots.length < 4) {
    slots = [...slots, createSlot()];
  }
}

function removeSlot(index: number) {
  if (slots.length > 2) {
    slots = slots.filter((_, i) => i !== index);
  }
}

let filledCount = $derived(slots.filter((s) => s.code.trim()).length);
let canCompare = $derived(filledCount >= 2 && !comparing);

async function compare() {
  const filledSlots = slots.filter((s) => s.code.trim());
  if (filledSlots.length < 2) return;

  comparing = true;
  hasCompared = true;

  for (const slot of slots) {
    if (slot.code.trim()) {
      slot.loading = true;
      slot.error = null;
      slot.result = null;
    }
  }

  const promises = slots.map(async (slot) => {
    if (!slot.code.trim()) return;
    const code = slot.code.trim().toUpperCase();
    try {
      slot.result = await LookupStock(code, slot.riskProfile);
    } catch (e: unknown) {
      slot.error = e instanceof Error ? e.message : String(e);
    } finally {
      slot.loading = false;
    }
  });

  await Promise.allSettled(promises);
  comparing = false;
}

let results = $derived(slots.map((s) => s.result));
let successCount = $derived(results.filter((r) => r !== null).length);
let totalFilled = $derived(slots.filter((s) => s.code.trim()).length);
</script>

<div class="mx-auto max-w-5xl px-4 py-8">
  <h1 class="text-2xl font-bold text-text-primary font-display">{t("comparison.title")}</h1>
  <p class="mt-1 mb-6 text-sm text-text-secondary">{t("comparison.subtitle")}</p>

  <!-- Ticker Inputs -->
  <form
    onsubmit={(e) => { e.preventDefault(); compare(); }}
    class="mb-6"
  >
    <div class="space-y-3">
      {#each slots as slot, i}
        <div class="flex items-center gap-2" data-testid="ticker-slot">
          <Input
            bind:value={slot.code}
            placeholder={t("comparison.tickerPlaceholder")}
            aria-label="Ticker {i + 1}"
            class="w-32 uppercase placeholder:normal-case placeholder:text-text-muted"
          />
          <Select
            bind:value={slot.riskProfile}
            aria-label="Risk profile {i + 1}"
            class="!w-auto"
          >
            <option value="CONSERVATIVE">{t("screener.conservative")}</option>
            <option value="MODERATE">{t("screener.moderate")}</option>
            <option value="AGGRESSIVE">{t("screener.aggressive")}</option>
          </Select>
          {#if watchlists.length > 0}
            <Select
              aria-label="Watchlist {i + 1}"
              class="!w-auto"
              onchange={async (e) => {
                const select = e.currentTarget;
                const value = select.value;
                if (!value) return;
                const [watchlistId, ticker] = value.split("|");
                if (ticker) {
                  slot.code = ticker;
                } else if (watchlistId) {
                  const items = await loadWatchlistItems(watchlistId);
                  if (items.length > 0) {
                    slot.code = items[0].ticker;
                  }
                }
                select.value = "";
              }}
            >
              <option value="">{t("comparison.fromWatchlist")}</option>
              {#each watchlists as wl (wl.id)}
                {#if watchlistItems.has(wl.id)}
                  {#each watchlistItems.get(wl.id) ?? [] as item (item.ticker)}
                    <option value="{wl.id}|{item.ticker}">{wl.name} - {item.ticker}</option>
                  {/each}
                {:else}
                  <option value={wl.id}>{wl.name}</option>
                {/if}
              {/each}
            </Select>
          {/if}
          {#if slot.loading}
            <LoaderCircle size={16} strokeWidth={2} class="animate-spin text-text-muted shrink-0" />
          {/if}
          {#if i >= 2}
            <button
              type="button"
              onclick={() => removeSlot(i)}
              class="rounded p-1.5 text-text-muted hover:bg-bg-tertiary hover:text-negative transition-fast focus-ring"
              aria-label={t("comparison.removeTicker")}
            >
              <Minus size={16} strokeWidth={2} />
            </button>
          {/if}
        </div>
        {#if slot.error}
          <div class="ml-0 rounded border border-negative/20 bg-negative-bg px-3 py-2 text-xs text-negative" role="alert">
            {slot.code.toUpperCase()}: {slot.error}
          </div>
        {/if}
      {/each}
    </div>

    <div class="mt-4 flex items-center gap-3">
      {#if slots.length < 4}
        <button
          type="button"
          onclick={addSlot}
          class="flex items-center gap-1.5 rounded border border-border-default px-3 py-2 text-sm text-text-secondary hover:bg-bg-tertiary transition-fast focus-ring"
        >
          <Plus size={14} strokeWidth={2} />
          {t("comparison.addTicker")}
        </button>
      {/if}
      <button
        type="submit"
        disabled={!canCompare}
        class="rounded bg-green-700 px-5 py-2 text-sm font-medium text-text-inverse hover:bg-green-800 disabled:opacity-50 focus-ring transition-fast"
      >
        {comparing ? t("comparison.comparing") : t("comparison.compare")}
      </button>
      {#if hasCompared && successCount < totalFilled}
        <span class="text-xs text-warning">
          {t("comparison.partialResults", { success: successCount, total: totalFilled })}
        </span>
      {/if}
    </div>
  </form>

  <!-- Empty State -->
  {#if !hasCompared}
    <EmptyState
      icon={ArrowLeftRight}
      title={t("comparison.emptyTitle")}
      description={t("comparison.emptyDescription")}
    />
  {/if}

  <!-- Comparison Table -->
  {#if hasCompared && successCount > 0}
    <ComparisonTable {results} />
  {/if}
</div>
