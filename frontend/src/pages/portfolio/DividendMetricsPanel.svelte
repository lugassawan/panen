<script lang="ts">
import { Coins } from "lucide-svelte";
import { getDividendIndicatorDisplay } from "../../lib/dividend-indicator";
import { formatDecimal, formatPercent, formatRupiah } from "../../lib/format";
import type { DividendMetricsResponse } from "../../lib/types";

let {
  ticker,
  dividendMetrics,
}: {
  ticker: string;
  dividendMetrics: DividendMetricsResponse;
} = $props();

const display = $derived(getDividendIndicatorDisplay(dividendMetrics.indicator));
</script>

<div
  class="rounded border border-border-default bg-bg-elevated p-4"
  data-testid="dividend-metrics-panel"
>
  <div class="mb-3 flex items-center justify-between">
    <div class="flex items-center gap-2">
      <Coins size={16} strokeWidth={2} class="text-text-muted" />
      <h4 class="text-sm font-semibold text-text-primary">{ticker}</h4>
    </div>
    <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium {display.bgClass} {display.colorClass}">
      <span aria-hidden="true">{display.icon}</span>
      {display.label}
    </span>
  </div>

  <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
    <div>
      <p class="text-xs text-text-muted">Annual DPS</p>
      <p class="font-mono text-sm text-text-primary">{formatRupiah(dividendMetrics.annualDPS)}</p>
    </div>
    <div>
      <p class="text-xs text-text-muted">Yield on Cost</p>
      <p class="font-mono text-sm text-text-primary">{formatPercent(dividendMetrics.yieldOnCost)}</p>
    </div>
    <div>
      <p class="text-xs text-text-muted">Projected YoC</p>
      <p class="font-mono text-sm text-text-secondary">{formatPercent(dividendMetrics.projectedYoC)}</p>
    </div>
    <div>
      <p class="text-xs text-text-muted">Portfolio Yield</p>
      <p class="font-mono text-sm text-text-secondary">{formatDecimal(dividendMetrics.portfolioYield)}%</p>
    </div>
  </div>
</div>
