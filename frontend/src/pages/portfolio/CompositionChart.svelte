<script lang="ts">
import { ArcElement, Chart, DoughnutController, Legend, Tooltip } from "chart.js";
import { PieChart } from "lucide-svelte";
import { t } from "../../i18n";
import { accentPalette, defaultChartOptions } from "../../lib/chartColors.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import { formatPercent, formatRupiah } from "../../lib/format";
import type { HoldingWeight, Mode, SectorWeight } from "../../lib/types";

Chart.register(ArcElement, DoughnutController, Legend, Tooltip);

interface Props {
  holdingWeights: HoldingWeight[];
  sectorWeights: SectorWeight[];
  portfolioMode: Mode;
}

let { holdingWeights, sectorWeights, portfolioMode }: Props = $props();

let showByHolding = $state(true);
let canvas: HTMLCanvasElement | undefined = $state();

let activeData = $derived(
  showByHolding
    ? holdingWeights.map((w) => ({ label: w.ticker, value: w.value, pct: w.pct }))
    : sectorWeights.map((w) => ({ label: w.sector, value: w.value, pct: w.pct })),
);

$effect(() => {
  if (!canvas || activeData.length === 0) return;

  void portfolioMode;
  const opts = defaultChartOptions();
  const colors = accentPalette(activeData.length);

  const chart = new Chart(canvas, {
    type: "doughnut",
    data: {
      labels: activeData.map((d) => d.label),
      datasets: [
        {
          data: activeData.map((d) => d.value),
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
        legend: {
          position: "bottom",
          labels: {
            ...opts.plugins?.legend?.labels,
            padding: 12,
          },
        },
        tooltip: {
          ...opts.plugins?.tooltip,
          callbacks: {
            label(ctx) {
              const d = activeData[ctx.dataIndex];
              return `${d.label}: ${formatRupiah(d.value)} (${formatPercent(d.pct)})`;
            },
          },
        },
      },
    },
  });

  return () => chart.destroy();
});
</script>

<div data-testid="composition-chart" class="rounded border border-border-default bg-bg-elevated p-4">
  <div class="mb-3 flex items-center justify-between">
    <p class="text-xs font-semibold uppercase tracking-wider text-text-muted">{t("chart.composition")}</p>
    {#if holdingWeights.length > 0}
      <div class="flex gap-1 rounded bg-bg-tertiary p-0.5 text-xs" role="group" aria-label={t("chart.composition")}>
        <button
          type="button"
          class="rounded px-2 py-0.5 transition-fast focus-ring {showByHolding ? 'bg-bg-elevated text-text-primary shadow-sm' : 'text-text-secondary'}"
          onclick={() => (showByHolding = true)}
          aria-pressed={showByHolding}
        >
          {t("chart.byHolding")}
        </button>
        <button
          type="button"
          class="rounded px-2 py-0.5 transition-fast focus-ring {!showByHolding ? 'bg-bg-elevated text-text-primary shadow-sm' : 'text-text-secondary'}"
          onclick={() => (showByHolding = false)}
          aria-pressed={!showByHolding}
        >
          {t("chart.bySector")}
        </button>
      </div>
    {/if}
  </div>

  {#if holdingWeights.length === 0}
    <EmptyState icon={PieChart} title={t("chart.noComposition")} description={t("chart.noCompositionDesc")} />
  {:else}
    <div class="relative" style="height: 280px">
      <canvas bind:this={canvas} aria-label={t("chart.portfolioComposition")}></canvas>
    </div>

    <table class="sr-only">
      <caption>{t("chart.portfolioComposition")}</caption>
      <thead>
        <tr><th>{t("chart.segment")}</th><th>{t("chart.value")}</th><th>{t("chart.weight")}</th></tr>
      </thead>
      <tbody>
        {#each activeData as d}
          <tr><td>{d.label}</td><td>{formatRupiah(d.value)}</td><td>{formatPercent(d.pct)}</td></tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
