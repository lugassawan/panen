<script lang="ts">
import { ACTION_LABELS } from "../../lib/action";
import { formatPercent, formatRupiah } from "../../lib/format";
import type { ActionType, SuggestionResponse } from "../../lib/types";

let {
  suggestion,
}: {
  suggestion: SuggestionResponse;
} = $props();

const isBuy = $derived(
  suggestion.action === "BUY" ||
    suggestion.action === "AVERAGE_DOWN" ||
    suggestion.action === "AVERAGE_UP",
);
</script>

<div
  class="rounded border border-green-700/20 bg-green-50 p-4 dark:border-green-700/30 dark:bg-green-950/20"
  data-testid="suggestion-card"
>
  <h4 class="mb-3 text-sm font-semibold text-green-700 dark:text-green-400">
    Trade Suggestion: {ACTION_LABELS[suggestion.action as ActionType]}
  </h4>

  <div class="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
    <div class="text-text-secondary">Ticker</div>
    <div class="font-mono font-medium text-right">{suggestion.ticker}</div>

    <div class="text-text-secondary">Lots</div>
    <div class="font-mono text-right">{suggestion.lots}</div>

    <div class="text-text-secondary">Price/Share</div>
    <div class="font-mono text-right">{formatRupiah(suggestion.pricePerShare)}</div>

    <div class="text-text-secondary">Gross {isBuy ? "Cost" : "Proceeds"}</div>
    <div class="font-mono text-right">{formatRupiah(suggestion.grossCost)}</div>

    <div class="text-text-secondary">Fee</div>
    <div class="font-mono text-right">{formatRupiah(suggestion.fee)}</div>

    {#if suggestion.tax > 0}
      <div class="text-text-secondary">Tax</div>
      <div class="font-mono text-right">{formatRupiah(suggestion.tax)}</div>
    {/if}

    <div class="text-text-secondary font-medium">Net {isBuy ? "Cost" : "Proceeds"}</div>
    <div class="font-mono font-medium text-right">{formatRupiah(suggestion.netCost)}</div>
  </div>

  {#if isBuy}
    <div
      class="mt-3 border-t border-green-700/10 pt-3 grid grid-cols-2 gap-x-6 gap-y-2 text-sm"
    >
      <div class="text-text-secondary">New Avg Price</div>
      <div class="font-mono text-right">{formatRupiah(suggestion.newAvgBuyPrice)}</div>

      <div class="text-text-secondary">New Position</div>
      <div class="font-mono text-right">
        {suggestion.newPositionLots} lots ({formatPercent(suggestion.newPositionPct)})
      </div>
    </div>
  {:else}
    <div
      class="mt-3 border-t border-green-700/10 pt-3 grid grid-cols-2 gap-x-6 gap-y-2 text-sm"
    >
      <div class="text-text-secondary">Capital Gain</div>
      <div
        class="font-mono text-right {suggestion.capitalGainPct >= 0
          ? 'text-profit'
          : 'text-loss'}"
      >
        {suggestion.capitalGainPct >= 0 ? "+" : ""}{formatPercent(suggestion.capitalGainPct)}
      </div>
    </div>
  {/if}
</div>
