<script lang="ts">
import { formatDecimal } from "../../lib/format";
import type { BandStats, StockValuationResponse } from "../../lib/types";

let {
  label,
  results,
  bandKey,
  valueKey,
}: {
  label: string;
  results: (StockValuationResponse | null)[];
  bandKey: "pbvBand" | "perBand";
  valueKey: "pbv" | "per";
} = $props();

function percentInRange(value: number, min: number, max: number): number {
  if (max === min) return 50;
  return Math.min(100, Math.max(0, ((value - min) / (max - min)) * 100));
}
</script>

<tr class="border-t border-border-default">
  <td class="py-2 pr-4 text-sm text-text-secondary whitespace-nowrap">{label}</td>
  {#each results as r}
    <td class="py-2 px-3 text-center">
      {#if r}
        {@const band = r[bandKey] as BandStats | undefined}
        {#if band}
          {@const pct = percentInRange(r[valueKey] as number, band.min, band.max)}
          <div class="flex flex-col items-center gap-1">
            <span class="text-xs font-mono text-text-secondary">{formatDecimal(r[valueKey] as number)}</span>
            <div class="relative w-full h-2 rounded-full bg-bg-tertiary">
              <div
                class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-green-700"
                style="left: {pct}%"
              ></div>
            </div>
            <div class="flex justify-between w-full text-[10px] text-text-muted font-mono">
              <span>{formatDecimal(band.min)}</span>
              <span>{formatDecimal(band.max)}</span>
            </div>
          </div>
        {:else}
          <span class="text-xs text-text-muted">--</span>
        {/if}
      {:else}
        <span class="text-xs text-text-muted">--</span>
      {/if}
    </td>
  {/each}
</tr>
