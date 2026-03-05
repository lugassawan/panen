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
import { GetPriceHistory } from "../../../wailsjs/go/backend/App";
import { accentPalette, chartColors, defaultChartOptions } from "../../lib/chartColors.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import Select from "../../lib/components/Select.svelte";
import { formatRupiah } from "../../lib/format";
import type { PricePointResponse, PriceRange } from "../../lib/types";

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
  tickers: string[];
}

let { tickers }: Props = $props();

let selectedTicker = $state("");
let range: PriceRange = $state("1Y");
let loading = $state(false);
let error = $state<string | null>(null);
let allPoints = $state<PricePointResponse[]>([]);
let canvas: HTMLCanvasElement | undefined = $state();

// Auto-select first ticker if only one
$effect(() => {
  if (tickers.length === 1) {
    selectedTicker = tickers[0];
  } else if (tickers.length > 0 && !tickers.includes(selectedTicker)) {
    selectedTicker = tickers[0];
  }
});

// Fetch data when ticker changes
$effect(() => {
  if (!selectedTicker) return;
  loading = true;
  error = null;
  GetPriceHistory(selectedTicker)
    .then((result) => {
      allPoints = result ?? [];
    })
    .catch((e) => {
      error = e instanceof Error ? e.message : String(e);
      allPoints = [];
    })
    .finally(() => {
      loading = false;
    });
});

const RANGE_MONTHS: Record<PriceRange, number | null> = {
  "1M": 1,
  "3M": 3,
  "6M": 6,
  "1Y": 12,
  ALL: null,
};

const RANGES: PriceRange[] = ["1M", "3M", "6M", "1Y", "ALL"];

let filteredPoints: PricePointResponse[] = $derived.by(() => {
  const months = RANGE_MONTHS[range];
  if (months === null) return allPoints;
  const cutoff = new Date();
  cutoff.setMonth(cutoff.getMonth() - months);
  const cutoffStr = cutoff.toISOString().slice(0, 10);
  return allPoints.filter((p) => p.date >= cutoffStr);
});

// Render chart
$effect(() => {
  if (!canvas || filteredPoints.length === 0) return;

  const colors = chartColors();
  const opts = defaultChartOptions();
  const accent = accentPalette(1)[0];

  const chart = new Chart(canvas, {
    type: "line",
    data: {
      labels: filteredPoints.map((p) => p.date),
      datasets: [
        {
          data: filteredPoints.map((p) => p.close),
          borderColor: accent,
          backgroundColor: `${accent}20`,
          borderWidth: 2,
          pointRadius: 0,
          pointHitRadius: 8,
          fill: true,
          tension: 0.1,
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
            title(items) {
              return items[0]?.label ?? "";
            },
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
            maxTicksLimit: 8,
            maxRotation: 0,
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

<div data-testid="price-history-chart" class="rounded border border-border-default bg-bg-elevated p-4">
  <div class="mb-3 flex items-center justify-between gap-3">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Price History</p>
    <div class="flex items-center gap-2">
      {#if tickers.length > 1}
        <div class="w-32">
          <Select
            bind:value={selectedTicker}
            aria-label="Select ticker"
          >
            {#each tickers as ticker}
              <option value={ticker}>{ticker}</option>
            {/each}
          </Select>
        </div>
      {:else if tickers.length === 1}
        <span class="text-sm font-mono font-medium text-text-primary">{tickers[0]}</span>
      {/if}

      {#if allPoints.length > 0}
        <div class="flex gap-1 rounded bg-bg-tertiary p-0.5 text-xs" role="group" aria-label="Time range">
          {#each RANGES as r}
            <button
              type="button"
              class="rounded px-2 py-0.5 transition-fast focus-ring {range === r ? 'bg-bg-elevated text-text-primary shadow-sm' : 'text-text-secondary'}"
              onclick={() => (range = r)}
              aria-pressed={range === r}
            >
              {r}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  {#if !selectedTicker}
    <EmptyState icon={TrendingUp} title="Select a ticker" description="Choose a holding to view its price history." />
  {:else if loading}
    <div class="flex items-center justify-center py-12">
      <p class="text-sm text-text-muted">Loading price history…</p>
    </div>
  {:else if error}
    <div class="rounded border border-border-default bg-bg-elevated p-6 text-center">
      <p class="text-sm text-loss">{error}</p>
    </div>
  {:else if filteredPoints.length === 0}
    <EmptyState icon={TrendingUp} title="No price data" description="No historical price data available for this ticker." />
  {:else}
    <div class="relative" style="height: 300px">
      <canvas bind:this={canvas} aria-label="Price history line chart for {selectedTicker}"></canvas>
    </div>
  {/if}
</div>
