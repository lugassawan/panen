<script lang="ts">
import { PackageOpen } from "lucide-svelte";
import Button from "../../lib/components/Button.svelte";
import EmptyState from "../../lib/components/EmptyState.svelte";
import Tooltip from "../../lib/components/Tooltip.svelte";
import { getDividendIndicatorDisplay } from "../../lib/dividend-indicator";
import { formatPercent, formatRupiah } from "../../lib/format";
import { calcPL } from "../../lib/portfolio";
import type { HoldingResponse } from "../../lib/types";
import { getVerdictDisplay } from "../../lib/verdict";

interface Props {
  holdings: HoldingResponse[];
  onChecklist: (ticker: string) => void;
}

let { holdings, onChecklist }: Props = $props();
</script>

{#if holdings.length === 0}
  <EmptyState
    icon={PackageOpen}
    title="No holdings yet"
    description="Add your first holding to this portfolio using the form below."
  />
{:else}
  <div class="overflow-x-auto rounded border border-border-default">
    <table class="w-full text-sm" aria-label="Holdings">
      <thead class="border-b border-border-default bg-bg-secondary">
        <tr>
          <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Ticker</th>
          <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Avg Buy Price</th>
          <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Lots</th>
          <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">Current Price</th>
          <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-text-muted">P/L %</th>
          <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">
            <Tooltip text="Stock valuation verdict based on Graham analysis and risk profile">
              <span class="underline decoration-dotted cursor-help">Verdict</span>
            </Tooltip>
          </th>
          <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-text-muted">Action</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-border-default">
        {#each holdings as holding}
          {@const pl = calcPL(holding.currentPrice, holding.avgBuyPrice)}
          {@const verdict = holding.verdict ? getVerdictDisplay(holding.verdict) : null}
          <tr class="hover:bg-bg-tertiary">
            <td class="px-4 py-3 font-medium">{holding.ticker}</td>
            <td class="px-4 py-3 text-right font-mono text-text-secondary">{formatRupiah(holding.avgBuyPrice)}</td>
            <td class="px-4 py-3 text-right text-text-secondary">{holding.lots}</td>
            <td class="px-4 py-3 text-right font-mono text-text-secondary">
              {holding.currentPrice != null ? formatRupiah(holding.currentPrice) : "\u2014"}
            </td>
            <td
              class="px-4 py-3 text-right font-mono {pl != null && pl >= 0 ? 'text-profit' : ''} {pl != null && pl < 0 ? 'text-loss' : ''}"
              data-testid="pl-{holding.ticker}"
            >
              {#if pl != null}
                {pl >= 0 ? "+" : ""}{formatPercent(pl)}
              {:else}
                &mdash;
              {/if}
            </td>
            <td class="px-4 py-3">
              {#if holding.dividendMetrics}
                {@const divDisplay = getDividendIndicatorDisplay(holding.dividendMetrics.indicator)}
                <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium {divDisplay.bgClass} {divDisplay.colorClass}">
                  <span aria-hidden="true">{divDisplay.icon}</span>
                  {divDisplay.label}
                </span>
              {:else if verdict}
                <span class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium {verdict.bgClass} {verdict.colorClass}">
                  <span aria-hidden="true">{verdict.icon}</span>
                  {verdict.label}
                </span>
              {:else}
                <span class="text-text-muted">&mdash;</span>
              {/if}
            </td>
            <td class="px-4 py-3">
              <Button
                variant="ghost"
                size="sm"
                onclick={() => onChecklist(holding.ticker)}
              >
                Checklist
              </Button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
