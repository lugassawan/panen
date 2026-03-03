<script lang="ts">
import { LoaderCircle, Plus, Trash2, X } from "lucide-svelte";
import {
  AddToWatchlist,
  CreateWatchlist,
  DeleteWatchlist,
  GetPresetItems,
  GetWatchlistItems,
  ListIndexNames,
  ListWatchlistSectors,
  ListWatchlists,
  RemoveFromWatchlist,
} from "../../wailsjs/go/backend/App";
import ConfirmDialog from "../components/ConfirmDialog.svelte";
import Badge from "../lib/components/Badge.svelte";
import Button from "../lib/components/Button.svelte";
import { formatPercent, formatRupiah } from "../lib/format";
import type { WatchlistItemResponse, WatchlistResponse } from "../lib/types";
import { getVerdictDisplay } from "../lib/verdict";

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
let newWatchlistName = $state("");
let createLoading = $state(false);
let createError = $state<string | null>(null);

// Delete watchlist confirm
let deletingWatchlist = $state<WatchlistResponse | null>(null);
let deleteLoading = $state(false);
let deleteError = $state<string | null>(null);

// Add ticker form
let addTickerInput = $state("");
let addTickerLoading = $state(false);
let addTickerError = $state<string | null>(null);

// Remove ticker
let removingTicker = $state<string | null>(null);

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
      const result = await GetPresetItems(activePreset);
      items = result ?? [];
    } else if (activeType === "watchlist" && activeWatchlist) {
      const result = await GetWatchlistItems(activeWatchlist.id, activeSector);
      items = result ?? [];
    }

    if (items.length > 0) {
      const sectorId = activeType === "preset" ? (activePreset ?? "") : (activeWatchlist?.id ?? "");
      const sectorList = await ListWatchlistSectors(sectorId, activeType === "preset");
      sectors = sectorList ?? [];
    }

    itemsState = "loaded";
  } catch (e: unknown) {
    itemsError = e instanceof Error ? e.message : String(e);
    itemsState = "error";
  }
}

async function reloadItemsWithSector() {
  if (!activeWatchlist && !activePreset) return;
  itemsState = "loading";
  itemsError = null;
  try {
    if (activeType === "preset" && activePreset) {
      const result = await GetPresetItems(activePreset);
      items = result ?? [];
    } else if (activeType === "watchlist" && activeWatchlist) {
      const result = await GetWatchlistItems(activeWatchlist.id, activeSector);
      items = result ?? [];
    }
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
  reloadItemsWithSector();
}

async function submitCreateWatchlist(e: Event) {
  e.preventDefault();
  const name = newWatchlistName.trim();
  if (!name) return;
  createLoading = true;
  createError = null;
  try {
    await CreateWatchlist(name);
    newWatchlistName = "";
    showCreateForm = false;
    await load();
  } catch (err: unknown) {
    createError = err instanceof Error ? err.message : String(err);
  } finally {
    createLoading = false;
  }
}

function startDeleteWatchlist(wl: WatchlistResponse) {
  deletingWatchlist = wl;
  deleteError = null;
}

async function confirmDeleteWatchlist() {
  if (!deletingWatchlist) return;
  deleteLoading = true;
  deleteError = null;
  try {
    await DeleteWatchlist(deletingWatchlist.id);
    if (activeWatchlist?.id === deletingWatchlist.id) {
      activeWatchlist = null;
      activeType = null;
      items = [];
      itemsState = "idle";
    }
    deletingWatchlist = null;
    await load();
  } catch (e: unknown) {
    deleteError = e instanceof Error ? e.message : String(e);
  } finally {
    deleteLoading = false;
  }
}

function cancelDeleteWatchlist() {
  deletingWatchlist = null;
  deleteError = null;
}

async function submitAddTicker(e: Event) {
  e.preventDefault();
  const ticker = addTickerInput.trim().toUpperCase();
  if (!ticker || !activeWatchlist) return;
  addTickerLoading = true;
  addTickerError = null;
  try {
    await AddToWatchlist(activeWatchlist.id, ticker);
    addTickerInput = "";
    await loadItems();
  } catch (err: unknown) {
    addTickerError = err instanceof Error ? err.message : String(err);
  } finally {
    addTickerLoading = false;
  }
}

async function removeTicker(ticker: string) {
  if (!activeWatchlist) return;
  removingTicker = ticker;
  try {
    await RemoveFromWatchlist(activeWatchlist.id, ticker);
    await loadItems();
  } catch {
    // silently fail — future: surface error inline
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
      <p class="mb-1.5 px-1 text-xs font-semibold uppercase tracking-wider text-text-muted">Preset Indices</p>
      {#if state === "loading"}
        <div class="flex items-center justify-center py-4 text-text-muted" role="status">
          <LoaderCircle size={16} strokeWidth={2} class="animate-spin" />
        </div>
      {:else if indexNames.length === 0}
        <p class="px-1 py-2 text-xs text-text-muted italic">No indices available</p>
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
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">My Watchlists</p>
        {#if !showCreateForm}
          <button
            type="button"
            class="rounded p-0.5 text-text-muted transition-fast focus-ring hover:bg-bg-tertiary hover:text-text-primary"
            onclick={() => { showCreateForm = true; newWatchlistName = ""; createError = null; }}
            aria-label="New Watchlist"
          >
            <Plus size={14} strokeWidth={2} />
          </button>
        {/if}
      </div>

      {#if showCreateForm}
        <form
          onsubmit={submitCreateWatchlist}
          class="mb-2 rounded border border-border-default bg-bg-elevated px-2 py-2"
        >
          <input
            bind:value={newWatchlistName}
            placeholder="Watchlist name"
            aria-label="New watchlist name"
            class="mb-1.5 w-full rounded border border-border-default bg-bg-primary px-2 py-1 text-xs text-text-primary placeholder:text-text-muted outline-none focus:border-green-700 focus-ring"
            disabled={createLoading}
          />
          {#if createError}
            <p class="mb-1.5 text-xs text-negative">{createError}</p>
          {/if}
          <div class="flex gap-1.5">
            <button
              type="submit"
              disabled={createLoading || !newWatchlistName.trim()}
              class="flex-1 rounded bg-green-700 px-2 py-1 text-xs font-medium text-text-inverse transition-fast focus-ring hover:bg-green-800 disabled:pointer-events-none disabled:opacity-50"
            >
              {createLoading ? "Adding…" : "Add"}
            </button>
            <button
              type="button"
              class="rounded px-2 py-1 text-xs text-text-secondary transition-fast focus-ring hover:bg-bg-tertiary"
              onclick={() => { showCreateForm = false; createError = null; }}
              disabled={createLoading}
            >
              Cancel
            </button>
          </div>
        </form>
      {/if}

      {#if state === "loading" && !showCreateForm}
        <div class="flex items-center justify-center py-4 text-text-muted" role="status">
          <LoaderCircle size={16} strokeWidth={2} class="animate-spin" />
        </div>
      {:else if state === "list" && watchlists.length === 0 && !showCreateForm}
        <p class="px-1 py-2 text-xs italic text-text-muted">No watchlists yet</p>
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
                onclick={(e) => { e.stopPropagation(); startDeleteWatchlist(wl); }}
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
          <p class="mb-1 font-medium text-text-primary">Select a watchlist or index</p>
          <p class="text-sm text-text-secondary">
            Choose a preset index or one of your watchlists from the left panel.
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
        <div class="border-b border-border-default px-6 py-3">
          <form onsubmit={submitAddTicker} class="flex items-center gap-2">
            <input
              bind:value={addTickerInput}
              placeholder="Add ticker (e.g. BBCA)"
              aria-label="Add ticker to watchlist"
              class="w-48 rounded border border-border-default bg-bg-elevated px-3 py-1.5 text-sm uppercase text-text-primary placeholder:normal-case placeholder:text-text-muted outline-none focus:border-green-700 focus-ring transition-fast"
              disabled={addTickerLoading}
            />
            <Button type="submit" size="sm" disabled={addTickerLoading || !addTickerInput.trim()} loading={addTickerLoading}>
              <Plus size={14} strokeWidth={2} />
              Add
            </Button>
            {#if addTickerError}
              <p class="text-sm text-negative">{addTickerError}</p>
            {/if}
          </form>
        </div>
      {/if}

      <!-- Items Content -->
      {#if itemsState === "loading"}
        <div class="flex flex-1 items-center justify-center gap-2 py-16 text-text-secondary" role="status">
          <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
          <span>Loading items…</span>
        </div>
      {:else if itemsState === "error"}
        <div class="mx-6 mt-4 rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
          {itemsError}
        </div>
      {:else if itemsState === "loaded"}
        <!-- Sector Filter Chips -->
        {#if sectors.length > 0}
          <div class="flex flex-wrap gap-2 border-b border-border-default px-6 py-3">
            <button
              type="button"
              class="rounded-full border px-3 py-1 text-xs font-medium transition-fast focus-ring {activeSector === ''
                ? 'border-green-700 bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                : 'border-border-default bg-bg-elevated text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
              onclick={() => selectSector("")}
            >
              All
            </button>
            {#each sectors as sector}
              <button
                type="button"
                class="rounded-full border px-3 py-1 text-xs font-medium transition-fast focus-ring {activeSector === sector
                  ? 'border-green-700 bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                  : 'border-border-default bg-bg-elevated text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
                onclick={() => selectSector(sector)}
              >
                {sector}
              </button>
            {/each}
          </div>
        {/if}

        {#if filteredItems.length === 0}
          <div class="flex flex-1 items-center justify-center py-16 text-center">
            <div>
              <p class="mb-1 font-medium text-text-primary">No items found</p>
              <p class="text-sm text-text-secondary">
                {#if activeType === "watchlist"}
                  Use the input above to add tickers to this watchlist.
                {:else}
                  This index has no available items.
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
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Ticker</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Sector</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Price</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">ROE</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">DER</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">EPS</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Div Yield</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Verdict</th>
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
                        <Badge variant={verdictBadgeVariant(item.verdict)}>
                          <span aria-hidden="true">{verdictDisplay.icon}</span>
                          {verdictDisplay.label}
                        </Badge>
                      {:else}
                        <span class="text-text-muted">\u2014</span>
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
  <ConfirmDialog
    title="Delete Watchlist"
    confirmLabel="Delete"
    confirmVariant="danger"
    loading={deleteLoading}
    onConfirm={confirmDeleteWatchlist}
    onCancel={cancelDeleteWatchlist}
  >
    <p>Are you sure you want to delete <strong>{deletingWatchlist.name}</strong>?</p>
    <p class="mt-1">This action cannot be undone.</p>
    {#if deleteError}
      <div class="mt-3 rounded border border-negative/20 bg-negative-bg px-3 py-2 text-sm text-negative" role="alert">
        {deleteError}
      </div>
    {/if}
  </ConfirmDialog>
{/if}
