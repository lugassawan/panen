<script lang="ts">
import { CheckCircle2, SearchX, SlidersHorizontal, XCircle } from "lucide-svelte";
import {
  ListScreenerIndices,
  ListScreenerSectors,
  RunScreen,
} from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import Alert from "../../lib/components/Alert.svelte";
import Badge from "../../lib/components/Badge.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import LoadingState from "../../lib/components/LoadingState.svelte";
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
            .map((s) => s.trim().toUpperCase())
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
    <h1 class="text-2xl font-display font-bold text-text-primary">{t("screener.title")}</h1>
    <p class="mt-1 text-sm text-text-secondary">{t("screener.subtitle")}</p>
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
    <EmptyState icon={SlidersHorizontal} title={t("screener.configurePrompt")} description={t("screener.configureDescription")} />
  {:else if state === "loading"}
    <LoadingState message={t("screener.screening")} class="flex-1 py-16" />
  {:else if state === "error"}
    <div class="mx-6 mt-4">
      <Alert variant="negative">
        {error}
        <button
          type="button"
          class="mt-2 block text-xs font-medium underline hover:opacity-80 focus-ring rounded"
          onclick={runScreen}
        >
          {t("common.retry")}
        </button>
      </Alert>
    </div>
  {:else if state === "results"}
    <!-- Summary -->
    <div class="flex items-center gap-4 border-b border-border-default px-6 py-3 text-sm text-text-secondary">
      <span>{t("screener.stocksScreened", { count: results.length })}</span>
      <span class="flex items-center gap-1 text-positive">
        <CheckCircle2 size={14} strokeWidth={2} />
        {passCount} {t("common.pass").toLowerCase()}
      </span>
      <span class="flex items-center gap-1 text-negative">
        <XCircle size={14} strokeWidth={2} />
        {failCount} {t("common.fail").toLowerCase()}
      </span>
    </div>

    {#if results.length === 0}
      <EmptyState icon={SearchX} title={t("screener.noResults")} description={t("screener.noResultsHint")} />
    {:else}
      <!-- Results Table -->
      <div class="flex-1 overflow-x-auto">
        <table class="w-full text-sm" aria-label="Screener results">
          <thead class="sticky top-0 border-b border-border-default bg-bg-secondary">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">
                <Tooltip text={t("screener.statusTooltip")}>
                  <span class="underline decoration-dotted cursor-help">{t("screener.status")}</span>
                </Tooltip>
              </th>
              {#each [
                { key: "ticker", label: t("screener.ticker") },
                { key: "sector", label: t("screener.sector") },
                { key: "price", label: t("screener.price") },
                { key: "roe", label: "ROE" },
                { key: "der", label: "DER" },
                { key: "dividendYield", label: "DY" },
                { key: "verdict", label: t("screener.verdict") },
                { key: "score", label: t("screener.score") },
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
                    <Badge variant="warning">{t("common.noData")}</Badge>
                  {:else if item.passed}
                    <Badge variant="profit">
                      <CheckCircle2 size={12} strokeWidth={2} />
                      {t("common.pass")}
                    </Badge>
                  {:else}
                    <Badge variant="loss">
                      <XCircle size={12} strokeWidth={2} />
                      {t("common.fail")}
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
