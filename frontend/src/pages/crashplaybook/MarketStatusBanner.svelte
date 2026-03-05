<script lang="ts">
import Badge from "../../lib/components/Badge.svelte";
import { formatDecimal } from "../../lib/format";
import type { MarketStatusResponse } from "../../lib/types";

let { market }: { market: MarketStatusResponse } = $props();

const conditionConfig: Record<
  string,
  { variant: "value" | "dividend" | "profit" | "loss" | "warning"; label: string }
> = {
  NORMAL: { variant: "value", label: "Normal" },
  ELEVATED: { variant: "warning", label: "Elevated" },
  CORRECTION: { variant: "loss", label: "Correction" },
  CRASH: { variant: "loss", label: "Crash" },
  RECOVERY: { variant: "profit", label: "Recovery" },
};

const config = $derived(conditionConfig[market.condition] ?? conditionConfig.NORMAL);
</script>

<div class="rounded-lg border border-border-default bg-bg-elevated p-4">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <h3 class="text-sm font-medium text-text-secondary">IHSG Market Status</h3>
      <Badge variant={config.variant}>{config.label}</Badge>
    </div>
    <div class="flex items-center gap-4 text-sm">
      <span class="font-mono font-medium text-text-primary">{formatDecimal(market.ihsgPrice)}</span>
      <span class="font-mono {market.drawdownPct < 0 ? 'text-loss' : 'text-profit'}">
        {formatDecimal(market.drawdownPct)}%
      </span>
    </div>
  </div>
  <div class="mt-2 flex items-center justify-between text-xs text-text-secondary">
    <span>Peak: <span class="font-mono">{formatDecimal(market.ihsgPeak)}</span></span>
    <span>Last fetched: {new Date(market.fetchedAt).toLocaleString("id-ID")}</span>
  </div>
</div>
