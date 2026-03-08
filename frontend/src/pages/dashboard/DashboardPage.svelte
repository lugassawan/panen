<script lang="ts">
import { ArcElement, Chart, DoughnutController, Legend, Tooltip } from "chart.js";
import { Landmark, TrendingDown, TrendingUp } from "lucide-svelte";
import { untrack } from "svelte";
import { GetDashboardOverview } from "../../../wailsjs/go/backend/App";
import { t } from "../../i18n";
import { accentPalette, defaultChartOptions } from "../../lib/chartColors.svelte";
import Alert from "../../lib/components/Alert.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import SkeletonCard from "../../lib/components/SkeletonCard.svelte";
import { formatDate, formatPercent, formatRupiah } from "../../lib/format";
import type { DashboardOverviewResponse, Page } from "../../lib/types";

Chart.register(ArcElement, DoughnutController, Legend, Tooltip);

interface Props {
  onNavigate: (page: Page) => void;
}

let { onNavigate }: Props = $props();

type State = "loading" | "ready" | "empty" | "error";

let state = $state<State>("loading");
let data = $state<DashboardOverviewResponse | null>(null);
let errorMsg = $state("");

let portfolioCanvas: HTMLCanvasElement | undefined = $state();
let sectorCanvas: HTMLCanvasElement | undefined = $state();

async function load() {
  state = "loading";
  try {
    const result = await GetDashboardOverview();
    data = result;
    state = !result.portfolios?.length ? "empty" : "ready";
  } catch (e) {
    errorMsg = String(e);
    state = "error";
  }
}

$effect(() => {
  untrack(() => load());
});

function renderChart(
  canvas: HTMLCanvasElement | undefined,
  items: { label: string; value: number; pct: number }[],
) {
  if (!canvas || items.length === 0) return undefined;
  const opts = defaultChartOptions();
  const colors = accentPalette(items.length);
  const chart = new Chart(canvas, {
    type: "doughnut",
    data: {
      labels: items.map((d) => d.label),
      datasets: [
        {
          data: items.map((d) => d.value),
          backgroundColor: colors,
          borderWidth: 2,
          borderColor: "transparent",
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: { duration: 200 },
      plugins: {
        legend: { position: "bottom", labels: { ...opts.plugins?.legend?.labels, padding: 12 } },
        tooltip: {
          ...opts.plugins?.tooltip,
          callbacks: {
            label: (ctx) =>
              `${items[ctx.dataIndex].label}: ${formatRupiah(items[ctx.dataIndex].value)} (${formatPercent(items[ctx.dataIndex].pct)})`,
          },
        },
      },
    },
  });
  return () => chart.destroy();
}

$effect(() => {
  if (state !== "ready" || !data) return;
  return renderChart(portfolioCanvas, data.portfolioAllocation);
});

$effect(() => {
  if (state !== "ready" || !data) return;
  return renderChart(sectorCanvas, data.sectorAllocation);
});

function plColor(amount: number): string {
  if (amount > 0) return "text-profit";
  if (amount < 0) return "text-loss";
  return "text-text-secondary";
}

function formatPL(amount: number, pct: number): string {
  const sign = amount >= 0 ? "+" : "";
  return `${sign}${formatRupiah(amount)} (${sign}${formatPercent(pct)})`;
}

function txnTypeBadge(type: string): string {
  switch (type) {
    case "BUY":
      return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
    case "SELL":
      return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
    case "DIVIDEND":
      return "bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400";
    default:
      return "bg-bg-tertiary text-text-secondary";
  }
}

function txnTypeLabel(type: string): string {
  switch (type) {
    case "BUY":
      return t("transactions.buy");
    case "SELL":
      return t("transactions.sell");
    case "DIVIDEND":
      return t("transactions.dividend");
    default:
      return type;
  }
}
</script>

<div class="mx-auto max-w-6xl space-y-6 p-6">
  <h1 class="text-2xl font-display font-bold text-text-primary">{t("dashboard.title")}</h1>

  {#if state === "loading"}
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
      {#each Array(3) as _}
        <SkeletonCard lines={2} />
      {/each}
    </div>
    <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
      {#each Array(2) as _}
        <SkeletonCard lines={6} />
      {/each}
    </div>
  {:else if state === "empty"}
    <EmptyState icon={Landmark} title={t("dashboard.emptyTitle")} description={t("dashboard.empty")}>
      {#snippet action()}
        <button
          class="rounded-lg bg-accent-primary px-4 py-2 text-sm font-medium text-white transition-fast hover:bg-accent-hover focus-ring"
          onclick={() => onNavigate("brokerage")}
        >
          {t("dashboard.goToBrokerage")}
        </button>
      {/snippet}
    </EmptyState>
  {:else if state === "error"}
    <Alert variant="negative">{errorMsg}</Alert>
  {:else if data}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
      <div class="rounded-lg border border-border-default bg-bg-elevated p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.totalMarketValue")}</p>
        <p class="mt-2 font-mono text-2xl font-bold text-text-primary">{formatRupiah(data.totalMarketValue)}</p>
      </div>
      <div class="rounded-lg border border-border-default bg-bg-elevated p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.totalPL")}</p>
        <p class="mt-2 font-mono text-2xl font-bold {plColor(data.totalPlAmount)}">
          {formatPL(data.totalPlAmount, data.totalPlPercent)}
        </p>
      </div>
      <div class="rounded-lg border border-border-default bg-bg-elevated p-5">
        <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.totalDividend")}</p>
        <p class="mt-2 font-mono text-2xl font-bold text-text-primary">{formatRupiah(data.totalDividendIncome)}</p>
      </div>
    </div>

    <!-- Allocation Charts -->
    {#if data.portfolioAllocation.length > 0 || data.sectorAllocation.length > 0}
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        {#if data.portfolioAllocation.length > 0}
          <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
            <p class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.portfolioAllocation")}</p>
            <div class="relative" style="height: 260px">
              <canvas bind:this={portfolioCanvas} aria-label={t("dashboard.portfolioAllocation")}></canvas>
            </div>
          </div>
        {/if}
        {#if data.sectorAllocation.length > 0}
          <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
            <p class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.sectorAllocation")}</p>
            <div class="relative" style="height: 260px">
              <canvas bind:this={sectorCanvas} aria-label={t("dashboard.sectorAllocation")}></canvas>
            </div>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Top Movers -->
    {#if data.topGainers.length > 0 || data.topLosers.length > 0}
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        {#if data.topGainers.length > 0}
          <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
            <div class="mb-3 flex items-center gap-2">
              <TrendingUp size={16} class="text-profit" />
              <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.topGainers")}</p>
            </div>
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border-default text-left text-xs text-text-muted">
                  <th class="pb-2">{t("dashboard.ticker")}</th>
                  <th class="pb-2">{t("dashboard.portfolio")}</th>
                  <th class="pb-2 text-right">{t("dashboard.plPct")}</th>
                  <th class="pb-2 text-right">{t("dashboard.plAmount")}</th>
                </tr>
              </thead>
              <tbody>
                {#each data.topGainers as h}
                  <tr class="border-b border-border-default last:border-0">
                    <td class="py-2 font-mono font-medium text-text-primary">{h.ticker}</td>
                    <td class="py-2 text-text-secondary">{h.portfolioName}</td>
                    <td class="py-2 text-right font-mono text-profit">+{formatPercent(h.plPercent)}</td>
                    <td class="py-2 text-right font-mono text-profit">+{formatRupiah(h.plAmount)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
        {#if data.topLosers.length > 0}
          <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
            <div class="mb-3 flex items-center gap-2">
              <TrendingDown size={16} class="text-loss" />
              <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.topLosers")}</p>
            </div>
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border-default text-left text-xs text-text-muted">
                  <th class="pb-2">{t("dashboard.ticker")}</th>
                  <th class="pb-2">{t("dashboard.portfolio")}</th>
                  <th class="pb-2 text-right">{t("dashboard.plPct")}</th>
                  <th class="pb-2 text-right">{t("dashboard.plAmount")}</th>
                </tr>
              </thead>
              <tbody>
                {#each data.topLosers as h}
                  <tr class="border-b border-border-default last:border-0">
                    <td class="py-2 font-mono font-medium text-text-primary">{h.ticker}</td>
                    <td class="py-2 text-text-secondary">{h.portfolioName}</td>
                    <td class="py-2 text-right font-mono text-loss">{formatPercent(h.plPercent)}</td>
                    <td class="py-2 text-right font-mono text-loss">{formatRupiah(h.plAmount)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Recent Transactions -->
    {#if data.recentTransactions.length > 0}
      <div class="rounded-lg border border-border-default bg-bg-elevated p-4">
        <p class="mb-3 text-xs font-semibold uppercase tracking-wider text-text-muted">{t("dashboard.recentTransactions")}</p>
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border-default text-left text-xs text-text-muted">
              <th class="pb-2">{t("transactions.type")}</th>
              <th class="pb-2">{t("dashboard.ticker")}</th>
              <th class="pb-2">{t("dashboard.portfolio")}</th>
              <th class="pb-2 text-right">{t("transactions.lots")}</th>
              <th class="pb-2 text-right">{t("transactions.total")}</th>
              <th class="pb-2 text-right">{t("transactions.date")}</th>
            </tr>
          </thead>
          <tbody>
            {#each data.recentTransactions as txn}
              <tr class="border-b border-border-default last:border-0">
                <td class="py-2">
                  <span class="inline-block rounded px-2 py-0.5 text-xs font-medium {txnTypeBadge(txn.type)}">{txnTypeLabel(txn.type)}</span>
                </td>
                <td class="py-2 font-mono font-medium text-text-primary">{txn.ticker}</td>
                <td class="py-2 text-text-secondary">{txn.portfolioName}</td>
                <td class="py-2 text-right font-mono text-text-secondary">{txn.lots}</td>
                <td class="py-2 text-right font-mono text-text-primary">{formatRupiah(txn.total)}</td>
                <td class="py-2 text-right text-text-muted">{formatDate(txn.date)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</div>
