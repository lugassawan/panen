<script lang="ts">
import { BarController, BarElement, CategoryScale, Chart, LinearScale, Tooltip } from "chart.js";
import { Wallet } from "lucide-svelte";
import { GetDividendIncomeSummary } from "../../../wailsjs/go/backend/App";
import { accentPalette, chartColors, defaultChartOptions } from "../../lib/chartColors.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import { formatRupiah } from "../../lib/format";
import type { DividendIncomeSummaryResponse } from "../../lib/types";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

interface Props {
  portfolioId: string;
}

let { portfolioId }: Props = $props();

let loading = $state(true);
let error = $state<string | null>(null);
let summary = $state<DividendIncomeSummaryResponse | null>(null);
let canvas: HTMLCanvasElement | undefined = $state();

const MONTH_LABELS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

$effect(() => {
  if (!portfolioId) return;
  loading = true;
  error = null;
  GetDividendIncomeSummary(portfolioId)
    .then((result) => {
      summary = result;
    })
    .catch((e) => {
      error = e instanceof Error ? e.message : String(e);
    })
    .finally(() => {
      loading = false;
    });
});

$effect(() => {
  if (!canvas || !summary || summary.monthlyBreakdown.length === 0) return;

  const colors = chartColors();
  const opts = defaultChartOptions();

  const data = new Array(12).fill(0);
  for (const m of summary.monthlyBreakdown) {
    data[m.month - 1] = m.amount;
  }

  const chart = new Chart(canvas, {
    type: "bar",
    data: {
      labels: MONTH_LABELS,
      datasets: [
        {
          data,
          backgroundColor: accentPalette(1)[0],
          borderRadius: 4,
        },
      ],
    },
    options: {
      ...opts,
      plugins: {
        ...opts.plugins,
        legend: { display: false },
        tooltip: {
          ...opts.plugins?.tooltip,
          callbacks: {
            label(ctx) {
              return formatRupiah(ctx.parsed.y);
            },
          },
        },
      },
      scales: {
        x: {
          ...opts.scales?.x,
          ticks: {
            ...((opts.scales?.x as Record<string, unknown>)?.ticks as Record<string, unknown>),
            font: { family: "DM Mono, monospace", size: 11 },
          },
        },
        y: {
          ...opts.scales?.y,
          ticks: {
            ...((opts.scales?.y as Record<string, unknown>)?.ticks as Record<string, unknown>),
            callback(value) {
              return formatRupiah(Number(value));
            },
          },
        },
      },
    },
  });

  return () => chart.destroy();
});
</script>

<div data-testid="dividend-income-chart" class="rounded border border-border-default bg-bg-elevated p-4">
  <div class="mb-3 flex items-center justify-between">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Dividend Income</p>
    {#if summary}
      <p class="text-sm font-mono font-medium text-gold-600">
        {formatRupiah(summary.totalAnnualIncome)}<span class="text-text-muted font-body">/yr</span>
      </p>
    {/if}
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <p class="text-sm text-text-muted">Loading income data…</p>
    </div>
  {:else if error}
    <div class="rounded border border-border-default bg-bg-elevated p-6 text-center">
      <p class="text-sm text-loss">{error}</p>
    </div>
  {:else if !summary || summary.monthlyBreakdown.length === 0}
    <EmptyState icon={Wallet} title="No income data" description="No dividend history available for holdings in this portfolio." />
  {:else}
    <div class="relative" style="height: 240px">
      <canvas bind:this={canvas} aria-label="Monthly dividend income bar chart"></canvas>
    </div>

    {#if summary.perStock.length > 0}
      <div class="mt-4">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-text-muted text-xs uppercase">
              <th class="text-left pb-2">Ticker</th>
              <th class="text-right pb-2">Income/yr</th>
              <th class="text-right pb-2">DY</th>
              <th class="text-right pb-2">Lots</th>
            </tr>
          </thead>
          <tbody>
            {#each summary.perStock as stock}
              <tr class="border-t border-border-default">
                <td class="py-1.5 font-mono font-medium">{stock.ticker}</td>
                <td class="py-1.5 text-right font-mono">{formatRupiah(stock.annualIncome)}</td>
                <td class="py-1.5 text-right font-mono">{stock.dividendYield.toFixed(2)}%</td>
                <td class="py-1.5 text-right font-mono">{stock.lots}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</div>
