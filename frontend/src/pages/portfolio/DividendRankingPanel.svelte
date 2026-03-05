<script lang="ts">
import { LoaderCircle, TrendingUp } from "lucide-svelte";
import { GetDividendRanking } from "../../../wailsjs/go/backend/App";
import { getDividendIndicatorDisplay } from "../../lib/dividend-indicator";
import { formatDecimal, formatPercent } from "../../lib/format";
import type { DividendRankItemResponse } from "../../lib/types";

let {
  portfolioId,
}: {
  portfolioId: string;
} = $props();

let loading = $state(true);
let error = $state<string | null>(null);
let items = $state<DividendRankItemResponse[]>([]);

async function loadRanking() {
  loading = true;
  error = null;
  try {
    const result = await GetDividendRanking(portfolioId);
    items = result ?? [];
  } catch (e: unknown) {
    error = e instanceof Error ? e.message : String(e);
  } finally {
    loading = false;
  }
}

loadRanking();
</script>

<div data-testid="dividend-ranking-panel">
  <div class="mb-3 flex items-center gap-2">
    <TrendingUp size={16} strokeWidth={2} class="text-text-muted" />
    <h3 class="text-xs font-semibold uppercase tracking-wider text-text-muted">DCA Ranking</h3>
  </div>

  {#if loading}
    <div class="flex items-center justify-center gap-2 py-6 text-text-secondary" role="status">
      <LoaderCircle size={16} strokeWidth={2} class="animate-spin" />
      <span class="text-sm">Loading ranking…</span>
    </div>
  {:else if error}
    <div class="rounded border border-negative/20 bg-negative-bg px-4 py-3 text-sm text-negative" role="alert">
      {error}
    </div>
  {:else if items.length === 0}
    <p class="py-4 text-center text-sm text-text-muted">No stocks to rank.</p>
  {:else}
    <div class="overflow-x-auto rounded border border-border-default">
      <table class="w-full text-sm" aria-label="Dividend Ranking">
        <thead class="border-b border-border-default bg-bg-secondary">
          <tr>
            <th class="px-3 py-2 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">#</th>
            <th class="px-3 py-2 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Ticker</th>
            <th class="px-3 py-2 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Indicator</th>
            <th class="px-3 py-2 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">DY</th>
            <th class="px-3 py-2 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">YoC</th>
            <th class="px-3 py-2 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Payout</th>
            <th class="px-3 py-2 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Weight</th>
            <th class="px-3 py-2 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Score</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border-default">
          {#each items as item, i}
            {@const display = getDividendIndicatorDisplay(item.indicator)}
            <tr class="hover:bg-bg-tertiary {item.isHolding ? '' : 'opacity-70'}">
              <td class="px-3 py-2 text-text-muted">{i + 1}</td>
              <td class="px-3 py-2 font-medium">
                {item.ticker}
                {#if !item.isHolding}
                  <span class="ml-1 text-xs text-text-muted">(watchlist)</span>
                {/if}
              </td>
              <td class="px-3 py-2">
                <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium {display.bgClass} {display.colorClass}">
                  <span aria-hidden="true">{display.icon}</span>
                  {display.label}
                </span>
              </td>
              <td class="px-3 py-2 text-right font-mono text-text-secondary">{formatPercent(item.dividendYield)}</td>
              <td class="px-3 py-2 text-right font-mono text-text-secondary">
                {item.yieldOnCost > 0 ? formatPercent(item.yieldOnCost) : "\u2014"}
              </td>
              <td class="px-3 py-2 text-right font-mono text-text-secondary">{formatPercent(item.payoutRatio, 0)}</td>
              <td class="px-3 py-2 text-right font-mono text-text-secondary">
                {item.positionPct > 0 ? formatPercent(item.positionPct) : "\u2014"}
              </td>
              <td class="px-3 py-2 text-right font-mono font-medium text-text-primary">{formatDecimal(item.score, 1)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
