<script lang="ts">
import { LoaderCircle, Plus, Trash2, X } from "lucide-svelte";
import {
  GetPresetItems,
  GetWatchlistItems,
  ListIndexNames,
  ListWatchlistSectors,
  ListWatchlists,
  RemoveFromWatchlist,
} from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Badge from "../../lib/components/Badge.svelte";
import LoadingState from "../../lib/components/LoadingState.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import { formatPercent, formatRupiah } from "../../lib/format";
import { toastStore } from "../../lib/stores/toast.svelte";
import type { WatchlistItemResponse, WatchlistResponse } from "../../lib/types";
import { getVerdictDisplay } from "../../lib/verdict";
import WatchlistAddTicker from "./WatchlistAddTicker.svelte";
import WatchlistCreateForm from "./WatchlistCreateForm.svelte";
import WatchlistDeleteDialog from "./WatchlistDeleteDialog.svelte";

type PageState = "loading" | "list" | "error";
type ItemsState = "idle" | "loading" | "loaded" | "error";
type ActiveType = "preset" | "watchlist" | null;

// Page state
let state = $state<PageState>("loading");
let error = $state<string | null>(null);

// Sidebar data
let watchlists = $state<WatchlistResponse[]>([]);
let indexNames = $state<string[]>([]);

// Selection
let activeType = $state<ActiveType>(null);
let activePreset = $state<string | null>(null);
let activeWatchlist = $state<WatchlistResponse | null>(null);

// Right panel items
let itemsState = $state<ItemsState>("idle");
let itemsError = $state<string | null>(null);
let items = $state<WatchlistItemResponse[]>([]);
let sectors = $state<string[]>([]);
let activeSector = $state<string>("");

// Create watchlist inline form
let showCreateForm = $state(false);

// Delete watchlist confirm
let deletingWatchlist = $state<WatchlistResponse | null>(null);

// Remove ticker
let removingTicker = $state<string | null>(null);
let removeError = $state<string | null>(null);

async function load() {
  state = "loading";
  error = null;
  try {
    const [wls, idxNames] = await Promise.all([ListWatchlists(), ListIndexNames()]);
    watchlists = wls ?? [];
    indexNames = idxNames ?? [];
    state = "list";
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

async function loadItems() {
  itemsState = "loading";
  itemsError = null;
  items = [];
  sectors = [];
  activeSector = "";

  try {
    if (activeType === "preset" && activePreset) {
      const result = await GetPresetItems(activePreset, "");
      items = result ?? [];
    } else if (activeType === "watchlist" && activeWatchlist) {
      const result = await GetWatchlistItems(activeWatchlist.id, "");
      items = result ?? [];
    }

    const sectorList = await ListWatchlistSectors();
    sectors = sectorList ?? [];

    itemsState = "loaded";
  } catch (e: unknown) {
    itemsError = e instanceof Error ? e.message : String(e);
    itemsState = "error";
  }
}

function selectPreset(name: string) {
  activeType = "preset";
  activePreset = name;
  activeWatchlist = null;
  activeSector = "";
  loadItems();
}

function selectWatchlist(wl: WatchlistResponse) {
  activeType = "watchlist";
  activeWatchlist = wl;
  activePreset = null;
  activeSector = "";
  loadItems();
}

function selectSector(sector: string) {
  activeSector = sector;
}

function handleWatchlistCreated() {
  showCreateForm = false;
  toastStore.add("Watchlist created", "success");
  load();
}

function handleWatchlistDeleted() {
  if (activeWatchlist?.id === deletingWatchlist?.id) {
    activeWatchlist = null;
    activeType = null;
    items = [];
    itemsState = "idle";
  }
  deletingWatchlist = null;
  toastStore.add("Watchlist deleted", "success");
  load();
}

async function removeTicker(ticker: string) {
  if (!activeWatchlist) return;
  removingTicker = ticker;
  removeError = null;
  try {
    await RemoveFromWatchlist(activeWatchlist.id, ticker);
    toastStore.add(`${ticker} removed`, "success");
    await loadItems();
  } catch (e: unknown) {
    removeError = e instanceof Error ? e.message : String(e);
  } finally {
    removingTicker = null;
  }
}

let activeHeaderName = $derived(
  activeType === "preset"
    ? (activePreset ?? "")
    : activeType === "watchlist"
      ? (activeWatchlist?.name ?? "")
      : "",
);

let filteredItems = $derived(
  activeSector ? items.filter((item) => item.sector === activeSector) : items,
);

function verdictBadgeVariant(verdict: string): "profit" | "loss" | "warning" {
  if (verdict === "UNDERVALUED") return "profit";
  if (verdict === "OVERVALUED") return "loss";
  return "warning";
}

load();
</script>

<div class="flex h-full">
  <!-- Left Panel -->
  <div class="flex w-56 flex-col overflow-y-auto border-r border-border-default bg-bg-secondary">
    <!-- Preset Indices Section -->
    <div class="px-3 pt-4 pb-2">
      <p class="mb-1.5 px-1 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("watchlist.presetIndices")}</p>
      {#if state === "loading"}
        <div class="flex items-center justify-center py-4 text-text-muted" role="status">
          <LoaderCircle size={16} strokeWidth={2} class="animate-spin" />
        </div>
      {:else if indexNames.length === 0}
        <p class="px-1 py-2 text-xs text-text-muted italic">{t("watchlist.noIndices")}</p>
      {:else}
        <ul role="list">
          {#each indexNames as name}
            <li>
              <button
                type="button"
                class="w-full rounded px-2 py-1.5 text-left text-sm transition-fast focus-ring {activeType === 'preset' && activePreset === name
                  ? 'bg-green-100 font-semibold text-green-800 dark:bg-green-900/30 dark:text-green-400'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
                onclick={() => selectPreset(name)}
              >
                {name}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    <div class="mx-3 border-t border-border-default"></div>

    <!-- My Watchlists Section -->
    <div class="flex-1 px-3 pt-3 pb-4">
      <div class="mb-1.5 flex items-center justify-between px-1">
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">{t("watchlist.myWatchlists")}</p>
        {#if !showCreateForm}
          <button
            type="button"
            class="rounded p-0.5 text-text-muted transition-fast focus-ring hover:bg-bg-tertiary hover:text-text-primary"
            onclick={() => { showCreateForm = true; }}
            aria-label={t("watchlist.newWatchlist")}
          >
            <Plus size={14} strokeWidth={2} />
          </button>
        {/if}
      </div>

      {#if showCreateForm}
        <WatchlistCreateForm
          onCreated={handleWatchlistCreated}
          onCancel={() => { showCreateForm = false; }}
        />
      {/if}

      {#if state === "loading" && !showCreateForm}
        <div class="flex items-center justify-center py-4 text-text-muted" role="status">
          <LoaderCircle size={16} strokeWidth={2} class="animate-spin" />
        </div>
      {:else if state === "list" && watchlists.length === 0 && !showCreateForm}
        <p class="px-1 py-2 text-xs italic text-text-muted">{t("watchlist.noWatchlists")}</p>
      {:else if state === "list"}
        <ul role="list">
          {#each watchlists as wl}
            <li class="group flex items-center gap-1">
              <button
                type="button"
                class="flex-1 truncate rounded px-2 py-1.5 text-left text-sm transition-fast focus-ring {activeType === 'watchlist' && activeWatchlist?.id === wl.id
                  ? 'bg-green-100 font-semibold text-green-800 dark:bg-green-900/30 dark:text-green-400'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
                onclick={() => selectWatchlist(wl)}
                title={wl.name}
              >
                {wl.name}
              </button>
              <button
                type="button"
                class="shrink-0 rounded p-0.5 text-text-muted opacity-0 transition-fast focus-ring group-hover:opacity-100 hover:bg-bg-tertiary hover:text-negative"
                onclick={(e) => { e.stopPropagation(); deletingWatchlist = wl; }}
                aria-label="Delete {wl.name}"
              >
                <Trash2 size={13} strokeWidth={2} />
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>

  <!-- Right Panel -->
  <div class="flex flex-1 flex-col overflow-y-auto bg-bg-primary">
    {#if activeType === null}
      <!-- Empty state: nothing selected -->
      <div class="flex flex-1 items-center justify-center py-24 text-center">
        <div>
          <p class="mb-1 font-medium text-text-primary">{t("watchlist.selectWatchlist")}</p>
          <p class="text-sm text-text-secondary">
            {t("watchlist.selectWatchlistDesc")}
          </p>
        </div>
      </div>
    {:else}
      <!-- Header -->
      <div class="border-b border-border-default px-6 py-4">
        <h2 class="text-lg font-semibold text-text-primary">{activeHeaderName}</h2>
      </div>

      <!-- Add Ticker (custom watchlists only) -->
      {#if activeType === "watchlist" && activeWatchlist}
        <WatchlistAddTicker watchlistId={activeWatchlist.id} onAdded={loadItems} />
      {/if}

      <!-- Items Content -->
      {#if itemsState === "loading"}
        <LoadingState message={t("watchlist.loadingItems")} class="flex-1 py-16" />
      {:else if itemsState === "error"}
        <div class="mx-6 mt-4 rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
          {itemsError}
        </div>
      {:else if itemsState === "loaded"}
        <!-- Sector Filter Chips -->
        {#if sectors.length > 0}
          <div class="flex flex-wrap gap-2 border-b border-border-default px-6 py-3" role="group" aria-label={t("watchlist.filterBySector")}>
            <button
              type="button"
              class="rounded-full border px-3 py-1 text-xs font-medium transition-fast focus-ring {activeSector === ''
                ? 'border-green-700 bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                : 'border-border-default bg-bg-elevated text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
              onclick={() => selectSector("")}
              aria-pressed={activeSector === ''}
            >
              {t("watchlist.allSectors")}
            </button>
            {#each sectors as sector}
              <button
                type="button"
                class="rounded-full border px-3 py-1 text-xs font-medium transition-fast focus-ring {activeSector === sector
                  ? 'border-green-700 bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                  : 'border-border-default bg-bg-elevated text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
                onclick={() => selectSector(sector)}
                aria-pressed={activeSector === sector}
              >
                {sector}
              </button>
            {/each}
          </div>
        {/if}

        {#if removeError}
          <div class="mx-6 mt-4 rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
            {t("watchlist.removeError", { error: removeError ?? "" })}
          </div>
        {/if}
        {#if filteredItems.length === 0}
          <div class="flex flex-1 items-center justify-center py-16 text-center">
            <div>
              <p class="mb-1 font-medium text-text-primary">{t("watchlist.noItems")}</p>
              <p class="text-sm text-text-secondary">
                {#if activeType === "watchlist"}
                  {t("watchlist.addTickerHint")}
                {:else}
                  {t("watchlist.indexEmptyHint")}
                {/if}
              </p>
            </div>
          </div>
        {:else}
          <!-- Items Table -->
          <div class="overflow-x-auto">
            <table class="w-full text-sm" aria-label="Watchlist items">
              <thead class="border-b border-border-default bg-bg-secondary">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">{t("watchlist.ticker")}</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">{t("watchlist.sector")}</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">{t("watchlist.price")}</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">ROE</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">DER</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">EPS</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Div Yield</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">{t("watchlist.verdict")}</th>
                  {#if activeType === "watchlist"}
                    <th class="px-4 py-3"></th>
                  {/if}
                </tr>
              </thead>
              <tbody class="divide-y divide-border-default">
                {#each filteredItems as item}
                  {@const verdictDisplay = item.verdict ? getVerdictDisplay(item.verdict) : null}
                  <tr class="hover:bg-bg-tertiary transition-fast">
                    <td class="px-4 py-3 font-medium font-mono text-text-primary">{item.ticker}</td>
                    <td class="px-4 py-3 text-text-secondary">
                      {item.sector || "\u2014"}
                    </td>
                    <td class="px-4 py-3 text-right font-mono text-text-secondary">
                      {item.price != null ? formatRupiah(item.price) : "\u2014"}
                    </td>
                    <td class="px-4 py-3 text-right font-mono text-text-secondary">
                      {item.roe != null ? formatPercent(item.roe) : "\u2014"}
                    </td>
                    <td class="px-4 py-3 text-right font-mono text-text-secondary">
                      {item.der != null ? item.der.toFixed(2) : "\u2014"}
                    </td>
                    <td class="px-4 py-3 text-right font-mono text-text-secondary">
                      {item.eps != null ? formatRupiah(item.eps) : "\u2014"}
                    </td>
                    <td class="px-4 py-3 text-right font-mono text-text-secondary">
                      {item.dividendYield != null ? formatPercent(item.dividendYield) : "\u2014"}
                    </td>
                    <td class="px-4 py-3">
                      {#if verdictDisplay && item.verdict}
                        <Tooltip text={verdictDisplay.description}>
                          <Badge variant={verdictBadgeVariant(item.verdict)}>
                            <span aria-hidden="true">{verdictDisplay.icon}</span>
                            {verdictDisplay.label}
                          </Badge>
                        </Tooltip>
                      {:else}
                        <span class="text-text-muted">&mdash;</span>
                      {/if}
                    </td>
                    {#if activeType === "watchlist"}
                      <td class="px-4 py-3">
                        <button
                          type="button"
                          class="rounded p-1 text-text-muted transition-fast focus-ring hover:bg-bg-tertiary hover:text-negative disabled:pointer-events-none disabled:opacity-50"
                          onclick={() => removeTicker(item.ticker)}
                          disabled={removingTicker === item.ticker}
                          aria-label="Remove {item.ticker}"
                        >
                          {#if removingTicker === item.ticker}
                            <LoaderCircle size={14} strokeWidth={2} class="animate-spin" />
                          {:else}
                            <X size={14} strokeWidth={2} />
                          {/if}
                        </button>
                      </td>
                    {/if}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      {:else}
        <!-- idle state: selected but not yet loaded (shouldn't normally show) -->
      {/if}
    {/if}
  </div>
</div>

{#if deletingWatchlist}
  <WatchlistDeleteDialog
    watchlist={deletingWatchlist}
    onDeleted={handleWatchlistDeleted}
    onCancel={() => { deletingWatchlist = null; }}
  />
{/if}
