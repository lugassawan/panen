<script lang="ts">
import {
  CategoryScale,
  Chart,
  Filler,
  LinearScale,
  LineController,
  LineElement,
  PointElement,
  Tooltip,
} from "chart.js";
import { TrendingUp } from "lucide-svelte";
import { GetYoCProgression } from "../../../wailsjs/go/backend/App";
import { accentPalette, defaultChartOptions } from "../../lib/chartColors.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import Select from "../../lib/components/Select.svelte";
import type { YoCPointResponse } from "../../lib/types";

Chart.register(
  CategoryScale,
  Filler,
  LinearScale,
  LineController,
  LineElement,
  PointElement,
  Tooltip,
);

interface Props {
  portfolioId: string;
  tickers: string[];
}

let { portfolioId, tickers }: Props = $props();

let selectedTicker = $state("");
let loading = $state(false);
let error = $state<string | null>(null);
let points = $state<YoCPointResponse[]>([]);
let canvas: HTMLCanvasElement | undefined = $state();

$effect(() => {
  if (tickers.length === 1) {
    selectedTicker = tickers[0];
  } else if (tickers.length > 0 && !tickers.includes(selectedTicker)) {
    selectedTicker = tickers[0];
  }
});

$effect(() => {
  if (!selectedTicker || !portfolioId) return;
  loading = true;
  error = null;
  GetYoCProgression(portfolioId, selectedTicker)
    .then((result) => {
      points = result ?? [];
    })
    .catch((e) => {
      error = e instanceof Error ? e.message : String(e);
      points = [];
    })
    .finally(() => {
      loading = false;
    });
});

$effect(() => {
  if (!canvas || points.length === 0) return;

  const opts = defaultChartOptions();

  const chart = new Chart(canvas, {
    type: "line",
    data: {
      labels: points.map((p) => p.date),
      datasets: [
        {
          data: points.map((p) => p.yoc),
          borderColor: accentPalette(1)[0],
          backgroundColor: `${accentPalette(1)[0]}20`,
          borderWidth: 2,
          pointRadius: 3,
          pointHitRadius: 8,
          fill: true,
          tension: 0.2,
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
              return `YoC: ${ctx.parsed.y.toFixed(2)}%`;
            },
          },
        },
      },
      scales: {
        x: {
          ...opts.scales?.x,
          ticks: {
            ...((opts.scales?.x as Record<string, unknown>)?.ticks as Record<string, unknown>),
            maxTicksLimit: 8,
            maxRotation: 0,
          },
        },
        y: {
          ...opts.scales?.y,
          ticks: {
            ...((opts.scales?.y as Record<string, unknown>)?.ticks as Record<string, unknown>),
            callback(value) {
              return `${Number(value).toFixed(1)}%`;
            },
          },
        },
      },
    },
  });

  return () => chart.destroy();
});
</script>

<div data-testid="yoc-progression-chart" class="rounded border border-border-default bg-bg-elevated p-4">
  <div class="mb-3 flex items-center justify-between gap-3">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Yield on Cost</p>
    {#if tickers.length > 1}
      <div class="w-32">
        <Select bind:value={selectedTicker} aria-label="Select ticker for YoC">
          {#each tickers as ticker}
            <option value={ticker}>{ticker}</option>
          {/each}
        </Select>
      </div>
    {:else if tickers.length === 1}
      <span class="text-sm font-mono font-medium text-text-primary">{tickers[0]}</span>
    {/if}
  </div>

  {#if !selectedTicker}
    <EmptyState icon={TrendingUp} title="Select a ticker" description="Choose a holding to view its yield on cost progression." />
  {:else if loading}
    <div class="flex items-center justify-center py-12">
      <p class="text-sm text-text-muted">Loading YoC data…</p>
    </div>
  {:else if error}
    <div class="rounded border border-border-default bg-bg-elevated p-6 text-center">
      <p class="text-sm text-loss">{error}</p>
    </div>
  {:else if points.length === 0}
    <EmptyState icon={TrendingUp} title="No YoC data" description="No dividend history available to compute yield on cost." />
  {:else}
    <div class="relative" style="height: 240px">
      <canvas bind:this={canvas} aria-label="Yield on cost progression chart for {selectedTicker}"></canvas>
    </div>
  {/if}
</div>
