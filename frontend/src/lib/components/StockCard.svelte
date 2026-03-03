<script lang="ts">
import { formatRupiah } from "../format";
import Badge from "./Badge.svelte";

let {
  ticker,
  name,
  price,
  change,
  changePercent,
  mode = "value",
  metrics = [],
}: {
  ticker: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
  mode?: "value" | "dividend";
  metrics?: { label: string; value: string; positive?: boolean }[];
} = $props();

const isPositive = $derived(change >= 0);
</script>

<div class="rounded-lg border border-border-default bg-bg-elevated p-4 shadow-xs">
  <div class="flex items-start justify-between">
    <div>
      <h3 class="font-display text-lg font-semibold">{ticker}</h3>
      <p class="text-sm text-text-secondary">{name}</p>
    </div>
    <Badge variant={mode}>{mode === "value" ? "\u{1F4C8} Value" : "\u{1F4B0} Dividend"}</Badge>
  </div>

  <div class="mt-3">
    <span class="font-mono text-xl font-semibold">
      {formatRupiah(price)}
    </span>
    <span class="ml-2 font-mono text-sm {isPositive ? 'text-profit' : 'text-loss'}">
      {isPositive ? "+" : ""}{formatRupiah(change)}
      ({isPositive ? "+" : ""}{changePercent.toFixed(2)}%)
    </span>
  </div>

  {#if metrics.length > 0}
    <div class="mt-3 grid grid-cols-3 gap-3">
      {#each metrics as metric}
        <div>
          <div class="text-xs font-semibold uppercase tracking-wide text-text-muted">
            {metric.label}
          </div>
          <div class="font-mono text-sm {metric.positive != null ? (metric.positive ? 'text-profit' : 'text-loss') : ''}">
            {metric.value}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
