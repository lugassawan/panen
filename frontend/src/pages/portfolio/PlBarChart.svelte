<script lang="ts">
import { BarController, BarElement, CategoryScale, Chart, LinearScale, Tooltip } from "chart.js";
import { BarChart } from "lucide-svelte";
import { chartColors, defaultChartOptions } from "../../lib/chartColors.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import { formatPercent, formatRupiah } from "../../lib/format";
import { calcPL, calcPLAbsolute } from "../../lib/portfolio";
import type { HoldingDetailResponse } from "../../lib/types";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

interface Props {
  holdings: HoldingDetailResponse[];
}

let { holdings }: Props = $props();

let showPercent = $state(true);
let canvas: HTMLCanvasElement | undefined = $state();

let validHoldings = $derived(
  holdings
    .filter((h) => h.currentPrice != null)
    .map((h) => ({
      ticker: h.ticker,
      plPct: calcPL(h.currentPrice, h.avgBuyPrice) ?? 0,
      plAbs: calcPLAbsolute(h) ?? 0,
    }))
    .sort((a, b) => b.plPct - a.plPct),
);

$effect(() => {
  if (!canvas || validHoldings.length === 0) return;

  const colors = chartColors();
  const opts = defaultChartOptions();
  const data = showPercent ? validHoldings.map((h) => h.plPct) : validHoldings.map((h) => h.plAbs);

  const barColors = data.map((v) => (v >= 0 ? colors.profit : colors.loss));

  const chart = new Chart(canvas, {
    type: "bar",
    data: {
      labels: validHoldings.map((h) => h.ticker),
      datasets: [
        {
          data,
          backgroundColor: barColors,
          borderRadius: 4,
        },
      ],
    },
    options: {
      ...opts,
      indexAxis: "y",
      plugins: {
        ...opts.plugins,
        legend: { display: false },
        tooltip: {
          ...opts.plugins?.tooltip,
          callbacks: {
            label(ctx) {
              const h = validHoldings[ctx.dataIndex];
              return `${formatPercent(h.plPct)} · ${formatRupiah(h.plAbs)}`;
            },
          },
        },
      },
      scales: {
        x: {
          ...opts.scales?.x,
          ticks: {
            ...((opts.scales?.x as Record<string, unknown>)?.ticks as Record<string, unknown>),
            callback(value) {
              return showPercent ? `${Number(value).toFixed(1)}%` : formatRupiah(Number(value));
            },
          },
        },
        y: {
          ...opts.scales?.y,
          ticks: {
            ...((opts.scales?.y as Record<string, unknown>)?.ticks as Record<string, unknown>),
            font: { family: "DM Sans, sans-serif", size: 12 },
          },
        },
      },
    },
  });

  return () => chart.destroy();
});
</script>

<div data-testid="pl-bar-chart" class="rounded border border-border-default bg-bg-elevated p-4">
  <div class="mb-3 flex items-center justify-between">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">P/L by Holding</p>
    {#if validHoldings.length > 0}
      <div class="flex gap-1 rounded bg-bg-tertiary p-0.5 text-xs" role="group" aria-label="P/L display mode">
        <button
          type="button"
          class="rounded px-2 py-0.5 transition-fast focus-ring {showPercent ? 'bg-bg-elevated text-text-primary shadow-sm' : 'text-text-secondary'}"
          onclick={() => (showPercent = true)}
          aria-pressed={showPercent}
        >
          %
        </button>
        <button
          type="button"
          class="rounded px-2 py-0.5 transition-fast focus-ring {!showPercent ? 'bg-bg-elevated text-text-primary shadow-sm' : 'text-text-secondary'}"
          onclick={() => (showPercent = false)}
          aria-pressed={!showPercent}
        >
          Rp
        </button>
      </div>
    {/if}
  </div>

  {#if validHoldings.length === 0}
    <EmptyState icon={BarChart} title="No P/L data" description="Holdings need current market prices to show P/L chart." />
  {:else}
    <div class="relative" style="height: {Math.max(validHoldings.length * 36, 120)}px">
      <canvas bind:this={canvas} aria-label="Profit and loss bar chart per holding"></canvas>
    </div>

    <!-- Screen reader accessible data table -->
    <table class="sr-only">
      <caption>P/L by holding</caption>
      <thead>
        <tr><th>Ticker</th><th>P/L %</th><th>P/L Rupiah</th></tr>
      </thead>
      <tbody>
        {#each validHoldings as h}
          <tr><td>{h.ticker}</td><td>{formatPercent(h.plPct)}</td><td>{formatRupiah(h.plAbs)}</td></tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
