<script lang="ts">
import { BarController, BarElement, CategoryScale, Chart, LinearScale, Tooltip } from "chart.js";
import { TrendingUp } from "lucide-svelte";
import { GetDGR } from "../../../wailsjs/go/backend/App";
import { chartColors, defaultChartOptions } from "../../lib/chartColors.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import Select from "../../lib/components/Select.svelte";
import type { DGRItemResponse } from "../../lib/types";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

interface Props {
  tickers: string[];
}

let { tickers }: Props = $props();

let selectedTicker = $state("");
let loading = $state(false);
let error = $state<string | null>(null);
let dgrData = $state<DGRItemResponse[]>([]);
let canvas: HTMLCanvasElement | undefined = $state();

$effect(() => {
  if (tickers.length === 1) {
    selectedTicker = tickers[0];
  } else if (tickers.length > 0 && !tickers.includes(selectedTicker)) {
    selectedTicker = tickers[0];
  }
});

$effect(() => {
  if (!selectedTicker) return;
  loading = true;
  error = null;
  GetDGR(selectedTicker)
    .then((result) => {
      dgrData = result ?? [];
    })
    .catch((e) => {
      error = e instanceof Error ? e.message : String(e);
      dgrData = [];
    })
    .finally(() => {
      loading = false;
    });
});

$effect(() => {
  if (!canvas || dgrData.length === 0) return;

  const colors = chartColors();
  const opts = defaultChartOptions();
  const growths = dgrData.map((d) => d.growthPct);
  const barColors = growths.map((v) => (v >= 0 ? colors.profit : colors.loss));

  const chart = new Chart(canvas, {
    type: "bar",
    data: {
      labels: dgrData.map((d) => String(d.year)),
      datasets: [
        {
          data: growths,
          backgroundColor: barColors,
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
            title(items) {
              const idx = items[0]?.dataIndex ?? 0;
              const d = dgrData[idx];
              return d ? `${d.year}` : "";
            },
            label(ctx) {
              const idx = ctx.dataIndex;
              const d = dgrData[idx];
              if (!d) return "";
              return `DPS: ${d.dps.toFixed(0)} · Growth: ${d.growthPct.toFixed(1)}%`;
            },
          },
        },
      },
      scales: {
        x: {
          ...opts.scales?.x,
        },
        y: {
          ...opts.scales?.y,
          ticks: {
            ...((opts.scales?.y as Record<string, unknown>)?.ticks as Record<string, unknown>),
            callback(value) {
              return `${Number(value).toFixed(0)}%`;
            },
          },
        },
      },
    },
  });

  return () => chart.destroy();
});
</script>

<div data-testid="dgr-chart" class="rounded border border-border-default bg-bg-elevated p-4">
  <div class="mb-3 flex items-center justify-between gap-3">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">Dividend Growth Rate</p>
    {#if tickers.length > 1}
      <div class="w-32">
        <Select bind:value={selectedTicker} aria-label="Select ticker for DGR">
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
    <EmptyState icon={TrendingUp} title="Select a ticker" description="Choose a holding to view its dividend growth rate." />
  {:else if loading}
    <div class="flex items-center justify-center py-12">
      <p class="text-sm text-text-muted">Loading DGR data…</p>
    </div>
  {:else if error}
    <div class="rounded border border-border-default bg-bg-elevated p-6 text-center">
      <p class="text-sm text-loss">{error}</p>
    </div>
  {:else if dgrData.length === 0}
    <EmptyState icon={TrendingUp} title="No DGR data" description="No dividend history available to compute growth rates." />
  {:else}
    <div class="relative" style="height: 240px">
      <canvas bind:this={canvas} aria-label="Dividend growth rate bar chart for {selectedTicker}"></canvas>
    </div>
  {/if}
</div>
