<script lang="ts">
import { CheckCircle2, LoaderCircle, XCircle } from "lucide-svelte";
import {
  ListScreenerIndices,
  ListScreenerSectors,
  RunScreen,
} from "../../../wailsjs/go/backend/App";
import Badge from "../../lib/components/Badge.svelte";
import SortableHeader from "../../lib/components/SortableHeader.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import { formatPercent, formatRupiah } from "../../lib/format";
import type { RiskProfile, ScreenerItemResponse } from "../../lib/types";
import { getVerdictDisplay } from "../../lib/verdict";
import ScreenerFilters from "./ScreenerFilters.svelte";

type PageState = "initial" | "loading" | "results" | "error";

let state = $state<PageState>("initial");
let error = $state<string | null>(null);
let results = $state<ScreenerItemResponse[]>([]);

// Filter state
let universeType = $state("INDEX");
let universeName = $state("");
let riskProfile = $state<RiskProfile>("MODERATE");
let sectorFilter = $state("");
let customTickers = $state("");

// Sort state
let sortField = $state("score");
let sortAsc = $state(false);

// Reference data
let indices = $state<string[]>([]);
let sectors = $state<string[]>([]);

async function loadReferenceData() {
  try {
    const [idx, sec] = await Promise.all([ListScreenerIndices(), ListScreenerSectors()]);
    indices = idx ?? [];
    sectors = sec ?? [];
  } catch {
    // Non-critical: selectors will be empty
  }
}

async function runScreen() {
  state = "loading";
  error = null;
  results = [];

  try {
    const tickers =
      universeType === "CUSTOM"
        ? customTickers
            .split(",")
            .map((t) => t.trim().toUpperCase())
            .filter(Boolean)
        : [];

    const res = await RunScreen(
      universeType,
      universeName,
      riskProfile,
      universeType === "SECTOR" ? "" : sectorFilter,
      "",
      false,
      tickers,
    );
    results = res ?? [];
    state = "results";
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
    state = "error";
  }
}

function toggleSort(field: string) {
  if (sortField === field) {
    sortAsc = !sortAsc;
  } else {
    sortField = field;
    sortAsc = false;
  }
}

function sortValue(item: ScreenerItemResponse, field: string): number | string {
  if (item.price == null) return -Infinity;
  switch (field) {
    case "ticker":
      return item.ticker;
    case "sector":
      return item.sector;
    case "price":
      return item.price ?? 0;
    case "roe":
      return item.roe ?? 0;
    case "der":
      return item.der ?? 0;
    case "dividendYield":
      return item.dividendYield ?? 0;
    case "verdict":
      return item.verdict ?? "";
    default:
      return item.score;
  }
}

function verdictBadgeVariant(verdict: string): "profit" | "loss" | "warning" {
  if (verdict === "UNDERVALUED") return "profit";
  if (verdict === "OVERVALUED") return "loss";
  return "warning";
}

let sortedResults = $derived(
  [...results].sort((a, b) => {
    const va = sortValue(a, sortField);
    const vb = sortValue(b, sortField);
    const cmp = va < vb ? -1 : va > vb ? 1 : 0;
    return sortAsc ? cmp : -cmp;
  }),
);

let passCount = $derived(results.filter((r) => r.passed).length);
let failCount = $derived(results.length - passCount);

loadReferenceData();
</script>

<div class="flex h-full flex-col bg-bg-primary">
  <!-- Page Header -->
  <div class="border-b border-border-default px-6 py-4">
    <h2 class="text-lg font-semibold text-text-primary">Stock Screener</h2>
    <p class="mt-0.5 text-sm text-text-secondary">Screen stocks against fundamental criteria by risk profile</p>
  </div>

  <!-- Filters -->
  <ScreenerFilters
    bind:universeType
    bind:universeName
    bind:riskProfile
    bind:sectorFilter
    bind:customTickers
    {indices}
    {sectors}
    loading={state === "loading"}
    onrun={runScreen}
  />

  <!-- Content -->
  {#if state === "initial"}
    <div class="flex flex-1 items-center justify-center py-24 text-center">
      <div>
        <p class="mb-1 font-medium text-text-primary">Configure and run a screen</p>
        <p class="text-sm text-text-secondary">
          Select a universe, choose a risk profile, and click "Run Screen" to discover stocks.
        </p>
      </div>
    </div>
  {:else if state === "loading"}
    <div class="flex flex-1 items-center justify-center gap-2 py-16 text-text-secondary" role="status">
      <LoaderCircle size={20} strokeWidth={2} class="animate-spin" />
      <span>Screening stocks...</span>
    </div>
  {:else if state === "error"}
    <div class="mx-6 mt-4 rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
      <p>{error}</p>
      <button
        type="button"
        class="mt-2 text-xs font-medium underline hover:opacity-80 focus-ring rounded"
        onclick={runScreen}
      >
        Retry
      </button>
    </div>
  {:else if state === "results"}
    <!-- Summary -->
    <div class="flex items-center gap-4 border-b border-border-default px-6 py-3 text-sm text-text-secondary">
      <span>{results.length} stock{results.length !== 1 ? "s" : ""} screened</span>
      <span class="flex items-center gap-1 text-positive">
        <CheckCircle2 size={14} strokeWidth={2} />
        {passCount} pass
      </span>
      <span class="flex items-center gap-1 text-negative">
        <XCircle size={14} strokeWidth={2} />
        {failCount} fail
      </span>
    </div>

    {#if results.length === 0}
      <div class="flex flex-1 items-center justify-center py-16 text-center">
        <div>
          <p class="mb-1 font-medium text-text-primary">No results</p>
          <p class="text-sm text-text-secondary">Try a different universe or broaden your filters.</p>
        </div>
      </div>
    {:else}
      <!-- Results Table -->
      <div class="flex-1 overflow-x-auto">
        <table class="w-full text-sm" aria-label="Screener results">
          <thead class="sticky top-0 border-b border-border-default bg-bg-secondary">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">
                <Tooltip text="Pass: meets all screening criteria for the selected risk profile. Fail: one or more criteria not met.">
                  <span class="underline decoration-dotted cursor-help">Status</span>
                </Tooltip>
              </th>
              {#each [
                { key: "ticker", label: "Ticker" },
                { key: "sector", label: "Sector" },
                { key: "price", label: "Price" },
                { key: "roe", label: "ROE" },
                { key: "der", label: "DER" },
                { key: "dividendYield", label: "Div Yield" },
                { key: "verdict", label: "Verdict" },
                { key: "score", label: "Score" },
              ] as col}
                <SortableHeader
                  label={col.label}
                  field={col.key}
                  currentSort={sortField}
                  ascending={sortAsc}
                  onclick={toggleSort}
                />
              {/each}
            </tr>
          </thead>
          <tbody class="divide-y divide-border-default">
            {#each sortedResults as item}
              {@const verdictDisplay = item.verdict ? getVerdictDisplay(item.verdict) : null}
              <tr class="transition-fast hover:bg-bg-tertiary">
                <td class="px-4 py-3">
                  {#if item.price == null}
                    <Badge variant="warning">No data</Badge>
                  {:else if item.passed}
                    <Badge variant="profit">
                      <CheckCircle2 size={12} strokeWidth={2} />
                      Pass
                    </Badge>
                  {:else}
                    <Badge variant="loss">
                      <XCircle size={12} strokeWidth={2} />
                      Fail
                    </Badge>
                  {/if}
                </td>
                <td class="px-4 py-3 font-mono font-medium text-text-primary">{item.ticker}</td>
                <td class="px-4 py-3 text-text-secondary">{item.sector || "\u2014"}</td>
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
                  {item.dividendYield != null ? formatPercent(item.dividendYield) : "\u2014"}
                </td>
                <td class="px-4 py-3">
                  {#if verdictDisplay && item.verdict}
                    <Badge variant={verdictBadgeVariant(item.verdict)}>
                      <span aria-hidden="true">{verdictDisplay.icon}</span>
                      {verdictDisplay.label}
                    </Badge>
                  {:else}
                    <span class="text-text-muted">&mdash;</span>
                  {/if}
                </td>
                <td class="px-4 py-3 text-right font-mono font-medium text-text-primary">
                  {item.score.toFixed(2)}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</div>
